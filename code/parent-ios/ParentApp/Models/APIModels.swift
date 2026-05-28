import Foundation

// MARK: - Family

struct CreateFamilyResponse: Decodable {
    let familyId: String
    let pairingCode: String
    let parentToken: String
}

// MARK: - Content

struct ImportItemRequest: Encodable {
    let familyId: String
    let sourceUrl: String
    let sourceType: String?
}

struct ImportItemResponse: Decodable {
    let id: String
    let processingStatus: String
    let isDuplicate: Bool
}

struct ItemResponse: Decodable {
    let id: String
    let familyId: String
    let title: String?
    let coverUrl: String?
    let summary: String?
    let sourceType: String
    let sourceUrl: String
    let processingStatus: String
    let playbackMode: String?
    let createdAt: String
    let updatedAt: String?
}

// MARK: - Source Types

enum SourceType: String, CaseIterable {
    case douyin = "douyin"
    case x = "x"
    case wechatChannels = "wechat_channels"
    case webArticle = "web_article"

    var displayName: String {
        switch self {
        case .douyin: return "抖音"
        case .x: return "X"
        case .wechatChannels: return "微信视频号"
        case .webArticle: return "网页"
        }
    }
}

// MARK: - Processing Status

enum ProcessingStatus: String {
    case pending = "pending"
    case ready = "ready"
    case failed = "failed"

    var displayName: String {
        switch self {
        case .pending: return "处理中"
        case .ready: return "已完成"
        case .failed: return "失败"
        }
    }
}
