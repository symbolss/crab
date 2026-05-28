import SwiftUI
import UniformTypeIdentifiers

/// Share Extension 入口视图
struct ShareView: View {
    @StateObject private var viewModel = ShareViewModel()
    @Environment(\.extensionContext) private var extensionContext

    var body: some View {
        NavigationView {
            Group {
                if viewModel.isLoading {
                    ProgressView("处理中...")
                } else if viewModel.isCompleted {
                    CompletedView(viewModel: viewModel)
                } else if let sharedURL = viewModel.sharedURL {
                    PreviewView(viewModel: viewModel, url: sharedURL)
                } else {
                    ErrorView(message: "无法识别分享的内容")
                }
            }
            .navigationTitle("分享到 Crab")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("取消") {
                        extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
                    }
                }
            }
            .alert("错误", isPresented: $viewModel.showError) {
                Button("确定", role: .cancel) { }
            } message: {
                Text(viewModel.errorMessage ?? "未知错误")
            }
        }
        .onAppear {
            viewModel.loadSharedItem(extensionContext: extensionContext)
        }
    }
}

// MARK: - Preview View

struct PreviewView: View {
    @ObservedObject var viewModel: ShareViewModel
    let url: URL

    var body: some View {
        VStack(spacing: 20) {
            // URL 预览
            VStack(alignment: .leading, spacing: 12) {
                Text("链接")
                    .font(.caption)
                    .foregroundColor(.secondary)

                Text(url.absoluteString)
                    .font(.body)
                    .lineLimit(3)
                    .padding()
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Color.gray.opacity(0.1))
                    .cornerRadius(8)

                // 来源类型
                HStack {
                    Text("来源:")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Text(viewModel.sourceType?.displayName ?? "网页")
                        .font(.caption)
                        .fontWeight(.medium)
                }
            }
            .padding()

            Spacer()

            // 发送按钮
            Button {
                Task {
                    await viewModel.sendToFamily()
                }
            } label: {
                if viewModel.isSending {
                    ProgressView()
                        .progressViewStyle(CircularProgressViewStyle(tint: .white))
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.blue)
                        .cornerRadius(10)
                } else {
                    Text("发送到家庭")
                        .font(.headline)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.blue)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                }
            }
            .disabled(viewModel.isSending)
            .padding(.horizontal)
        }
        .padding()
    }
}

// MARK: - Completed View

struct CompletedView: View {
    @ObservedObject var viewModel: ShareViewModel
    @Environment(\.extensionContext) private var extensionContext

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 60))
                .foregroundColor(.green)

            Text("已发送")
                .font(.title)
                .fontWeight(.bold)

            if viewModel.isDuplicate {
                Text("该内容已存在")
                    .font(.subheadline)
                    .foregroundColor(.orange)
            } else {
                Text("内容已成功发送到家庭")
                    .font(.subheadline)
                    .foregroundColor(.secondary)
            }

            Button("完成") {
                extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
            }
            .font(.headline)
            .frame(maxWidth: .infinity)
            .padding()
            .background(Color.blue)
            .foregroundColor(.white)
            .cornerRadius(10)
            .padding(.horizontal)
        }
        .padding()
    }
}

// MARK: - Error View

struct ErrorView: View {
    let message: String
    @Environment(\.extensionContext) private var extensionContext

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: "exclamationmark.triangle.fill")
                .font(.system(size: 60))
                .foregroundColor(.orange)

            Text("无法分享")
                .font(.title)
                .fontWeight(.bold)

            Text(message)
                .font(.subheadline)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)

            Button("关闭") {
                extensionContext?.completeRequest(returningItems: nil, completionHandler: nil)
            }
            .font(.headline)
            .frame(maxWidth: .infinity)
            .padding()
            .background(Color.gray)
            .foregroundColor(.white)
            .cornerRadius(10)
            .padding(.horizontal)
        }
        .padding()
    }
}

// MARK: - Share View Model

@MainActor
class ShareViewModel: ObservableObject {
    @Published var isLoading = true
    @Published var isSending = false
    @Published var isCompleted = false
    @Published var isDuplicate = false
    @Published var sharedURL: URL?
    @Published var sourceType: SourceType?
    @Published var showError = false
    @Published var errorMessage: String?

    private let storage = SharedStorage.shared
    private let apiClient = APIClient.shared

    func loadSharedItem(extensionContext: NSExtensionContext?) {
        guard let extensionContext = extensionContext else {
            errorMessage = "无法获取分享内容"
            showError = true
            isLoading = false
            return
        }

        guard let item = extensionContext.inputItems.first as? NSExtensionItem,
              let itemProviders = item.attachments else {
            errorMessage = "没有找到分享内容"
            showError = true
            isLoading = false
            return
        }

        // 查找 URL
        for provider in itemProviders {
            if provider.hasItemConformingToTypeIdentifier(UTType.url.identifier) {
                provider.loadItem(forTypeIdentifier: UTType.url.identifier, options: nil) { [weak self] data, error in
                    DispatchQueue.main.async {
                        if let url = data as? URL {
                            self?.sharedURL = url
                            self?.sourceType = self?.detectSourceType(from: url)
                        } else if let urlString = data as? String, let url = URL(string: urlString) {
                            self?.sharedURL = url
                            self?.sourceType = self?.detectSourceType(from: url)
                        } else {
                            self?.errorMessage = "无法解析链接"
                            self?.showError = true
                        }
                        self?.isLoading = false
                    }
                }
                return
            }
        }

        // 没有找到 URL
        errorMessage = "请分享一个网页链接"
        showError = true
        isLoading = false
    }

    func sendToFamily() async {
        guard let url = sharedURL,
              let familyId = storage.familyId else {
            errorMessage = "请先创建家庭"
            showError = true
            return
        }

        isSending = true
        errorMessage = nil

        do {
            let response = try await apiClient.importItem(
                familyId: familyId,
                sourceUrl: url.absoluteString,
                sourceType: sourceType?.rawValue
            )
            isDuplicate = response.isDuplicate
            isCompleted = true
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }

        isSending = false
    }

    private func detectSourceType(from url: URL) -> SourceType {
        let host = url.host?.lowercased() ?? ""

        if host.contains("douyin.com") || host.contains("v.douyin.com") {
            return .douyin
        } else if host.contains("twitter.com") || host.contains("x.com") {
            return .x
        } else if host.contains("channels.weixin.qq.com") || host.contains("weixin.qq.com") {
            return .wechatChannels
        } else {
            return .webArticle
        }
    }
}
