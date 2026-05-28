import SwiftUI

struct ContentView: View {
    @StateObject private var viewModel = ContentViewModel()

    var body: some View {
        NavigationView {
            Group {
                if viewModel.isLoading {
                    ProgressView("加载中...")
                } else if let familyId = viewModel.familyId {
                    FamilyView(viewModel: viewModel, familyId: familyId)
                } else {
                    SetupView(viewModel: viewModel)
                }
            }
            .navigationTitle("Crab")
            .alert("错误", isPresented: $viewModel.showError) {
                Button("确定", role: .cancel) { }
            } message: {
                Text(viewModel.errorMessage ?? "未知错误")
            }
        }
        .onAppear {
            viewModel.checkExistingFamily()
        }
    }
}

// MARK: - Setup View

struct SetupView: View {
    @ObservedObject var viewModel: ContentViewModel

    var body: some View {
        VStack(spacing: 24) {
            Image(systemName: "house.fill")
                .font(.system(size: 60))
                .foregroundColor(.blue)

            Text("欢迎使用 Crab")
                .font(.title)
                .fontWeight(.bold)

            Text("创建一个家庭，开始为孩子分享内容")
                .font(.subheadline)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)

            Button {
                Task {
                    await viewModel.createFamily()
                }
            } label: {
                Text("创建家庭")
                    .font(.headline)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.blue)
                    .foregroundColor(.white)
                    .cornerRadius(10)
            }
            .padding(.horizontal)
            .disabled(viewModel.isCreatingFamily)

            if viewModel.isCreatingFamily {
                ProgressView()
                    .padding(.top)
            }
        }
        .padding()
    }
}

// MARK: - Family View

struct FamilyView: View {
    @ObservedObject var viewModel: ContentViewModel
    let familyId: String

    var body: some View {
        VStack(spacing: 20) {
            // 配对码
            if let pairingCode = viewModel.pairingCode {
                VStack(spacing: 8) {
                    Text("家庭配对码")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Text(pairingCode)
                        .font(.system(.title, design: .monospaced))
                        .fontWeight(.bold)
                        .padding()
                        .background(Color.gray.opacity(0.1))
                        .cornerRadius(8)
                }
                .padding()
            }

            // 操作按钮
            VStack(spacing: 12) {
                Button {
                    UIPasteboard.general.string = viewModel.pairingCode
                } label: {
                    Label("复制配对码", systemImage: "doc.on.doc")
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.green.opacity(0.1))
                        .foregroundColor(.green)
                        .cornerRadius(10)
                }

                Button {
                    if let url = URL(string: "shareextension://") {
                        UIApplication.shared.open(url)
                    }
                } label: {
                    Label("分享内容", systemImage: "square.and.arrow.up")
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.blue)
                        .foregroundColor(.white)
                        .cornerRadius(10)
                }
            }
            .padding(.horizontal)

            Spacer()

            // 已分享内容
            if !viewModel.items.isEmpty {
                VStack {
                    Text("已分享 \(viewModel.items.count) 条内容")
                        .font(.subheadline)
                        .foregroundColor(.secondary)

                    List(viewModel.items) { item in
                        ItemRowView(item: item)
                    }
                }
            }
        }
        .padding()
        .task {
            await viewModel.loadItems()
        }
    }
}

// MARK: - Item Row View

struct ItemRowView: View {
    let item: ItemResponse

    var body: some View {
        HStack {
            AsyncImage(url: URL(string: item.coverUrl ?? "")) { image in
                image.resizable()
            } placeholder: {
                Color.gray.opacity(0.2)
            }
            .frame(width: 60, height: 60)
            .cornerRadius(8)

            VStack(alignment: .leading, spacing: 4) {
                Text(item.title ?? "无标题")
                    .font(.headline)
                    .lineLimit(2)

                Text(sourceTypeDisplayName)
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
    }

    private var sourceTypeDisplayName: String {
        SourceType(rawValue: item.sourceType)?.displayName ?? item.sourceType
    }
}

// MARK: - View Model

@MainActor
class ContentViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var isCreatingFamily = false
    @Published var familyId: String?
    @Published var pairingCode: String?
    @Published var items: [ItemResponse] = []
    @Published var showError = false
    @Published var errorMessage: String?

    private let storage = SharedStorage.shared
    private let apiClient = APIClient.shared

    func checkExistingFamily() {
        isLoading = true
        familyId = storage.familyId
        pairingCode = storage.pairingCode
        isLoading = false
    }

    func createFamily() async {
        isCreatingFamily = true
        errorMessage = nil

        do {
            let response = try await apiClient.createFamily()
            storage.familyId = response.familyId
            storage.pairingCode = response.pairingCode
            storage.jwtToken = response.parentToken

            familyId = response.familyId
            pairingCode = response.pairingCode
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }

        isCreatingFamily = false
    }

    func loadItems() async {
        guard let familyId = familyId else { return }

        isLoading = true

        do {
            items = try await apiClient.getItems(familyId: familyId)
        } catch {
            errorMessage = error.localizedDescription
            showError = true
        }

        isLoading = false
    }
}

#Preview {
    ContentView()
}
