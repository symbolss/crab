import Foundation

/// API 客户端，负责与后端通信
actor APIClient {
    static let shared = APIClient()

    // TODO: 配置实际的 API 基础 URL
    private let baseURL = "http://localhost:8080"

    private init() {}

    // MARK: - Family

    /// 创建家庭
    func createFamily() async throws -> CreateFamilyResponse {
        let url = URL(string: "\(baseURL)/api/family/create")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        guard httpResponse.statusCode == 200 else {
            throw APIError.httpError(httpResponse.statusCode)
        }

        return try JSONDecoder().decode(CreateFamilyResponse.self, from: data)
    }

    // MARK: - Content

    /// 导入内容到家庭
    func importItem(familyId: String, sourceUrl: String, sourceType: String? = nil) async throws -> ImportItemResponse {
        let url = URL(string: "\(baseURL)/api/items/import")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        // 添加 JWT 认证
        if let token = SharedStorage.shared.jwtToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let body: [String: Any?] = [
            "familyId": familyId,
            "sourceUrl": sourceUrl,
            "sourceType": sourceType
        ]
        request.httpBody = try? JSONSerialization.data(withJSONObject: body, options: [.omitNulls])

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        guard httpResponse.statusCode == 200 else {
            if httpResponse.statusCode == 401 {
                throw APIError.unauthorized
            }
            throw APIError.httpError(httpResponse.statusCode)
        }

        return try JSONDecoder().decode(ImportItemResponse.self, from: data)
    }

    /// 获取家庭内容列表
    func getItems(familyId: String) async throws -> [ItemResponse] {
        var components = URLComponents(string: "\(baseURL)/api/items")!
        components.queryItems = [URLQueryItem(name: "familyId", value: familyId)]

        var request = URLRequest(url: components.url!)
        request.httpMethod = "GET"

        if let token = SharedStorage.shared.jwtToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        guard httpResponse.statusCode == 200 else {
            throw APIError.httpError(httpResponse.statusCode)
        }

        return try JSONDecoder().decode([ItemResponse].self, from: data)
    }
}

// MARK: - Errors

enum APIError: Error, LocalizedError {
    case invalidResponse
    case httpError(Int)
    case unauthorized
    case decodingError(Error)

    var errorDescription: String? {
        switch self {
        case .invalidResponse:
            return "无效的响应"
        case .httpError(let code):
            return "HTTP 错误: \(code)"
        case .unauthorized:
            return "未授权，请重新登录"
        case .decodingError(let error):
            return "解析错误: \(error.localizedDescription)"
        }
    }
}

// MARK: - JSON Encoding Options

private extension JSONSerialization {
    static let omitNulls: JSONSerialization.WritingOptions = []
}
