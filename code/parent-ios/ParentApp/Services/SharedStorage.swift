import Foundation

/// 共享存储服务，使用 App Groups 在主应用和 Share Extension 之间共享数据
class SharedStorage {
    static let shared = SharedStorage()

    // App Groups ID，需要在 Xcode 中配置
    private let appGroupIdentifier = "group.com.family.crab"

    private var userDefaults: UserDefaults? {
        UserDefaults(suiteName: appGroupIdentifier)
    }

    private init() {}

    // MARK: - JWT Token

    var jwtToken: String? {
        get { userDefaults?.string(forKey: "jwtToken") }
        set { userDefaults?.set(newValue, forKey: "jwtToken") }
    }

    // MARK: - Family ID

    var familyId: String? {
        get { userDefaults?.string(forKey: "familyId") }
        set { userDefaults?.set(newValue, forKey: "familyId") }
    }

    // MARK: - Pairing Code

    var pairingCode: String? {
        get { userDefaults?.string(forKey: "pairingCode") }
        set { userDefaults?.set(newValue, forKey: "pairingCode") }
    }

    // MARK: - Clear

    func clearAll() {
        jwtToken = nil
        familyId = nil
        pairingCode = nil
    }
}
