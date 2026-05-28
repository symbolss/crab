# iOS 家长端技术方案

## 1. 技术选型

### 推荐方案: 原生 Swift + SwiftUI

| 维度 | 选择 | 理由 |
|------|------|------|
| 语言 | Swift 5.9+ | 系统级支持，Share Extension 稳定 |
| UI 框架 | SwiftUI | 现代化声明式 UI，开发效率高 |
| 数据持久化 | SwiftData | Apple 原生，与 SwiftUI 无缝集成 |
| 网络层 | URLSession + async/await | 原生支持，无需第三方依赖 |
| 最低版本 | iOS 17.0+ | SwiftData 需要，API 稳定 |

### 方案对比

| 方案 | 优点 | 缺点 | 推荐 |
|------|------|------|------|
| **原生 Swift** | Share Extension 成熟稳定、系统级支持、内存占用小 | 仅支持 iOS | ⭐⭐⭐⭐⭐ |
| Flutter | 跨平台代码复用 | Share Extension 有已知兼容问题、插件冲突、调试困难 | ⭐⭐ |

### Flutter Share Extension 已知问题
- [Issue #107482](https://github.com/flutter/flutter/issues/107482): 官方支持不完善
- [Issue #168263](https://github.com/flutter/flutter/issues/168263): GeneratedPluginRegistrant 找不到
- 需要原生桥接，增加复杂度

## 2. 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      iOS Parent App                          │
├────────────────────────┬────────────────────────────────────┤
│       Main App         │         Share Extension            │
├────────────────────────┼────────────────────────────────────┤
│ • 首次启动创建家庭     │ • 接收系统分享的链接              │
│ • 显示配对码           │ • 解析 URL 和标题                 │
│ • 孩子列表管理         │ • 展示最小预览                    │
│ • 设置（时长限制）     │ • 选择目标孩子                    │
│                        │ • 调用后端 API 发送               │
├────────────────────────┴────────────────────────────────────┤
│                    App Groups (共享容器)                     │
│  • UserDefaults: JWT Token, Family ID                       │
│  • Keychain: 敏感数据 (可选)                                 │
├─────────────────────────────────────────────────────────────┤
│                       Go Backend API                         │
│  POST /api/family/create    POST /api/items/import          │
│  POST /api/family/pair      POST /api/items/:id/assign      │
│  GET  /api/children                                       │
└─────────────────────────────────────────────────────────────┘
```

## 3. Share Extension 实现细节

### 3.1 创建 Share Extension

1. 在 Xcode 中: File → New → Target → Share Extension
2. 配置 App Groups (主 App 和 Extension 必须共享)
3. 删除默认 Storyboard，使用 SwiftUI

### 3.2 Info.plist 配置

```xml
<key>NSExtension</key>
<dict>
    <key>NSExtensionAttributes</key>
    <dict>
        <key>NSExtensionActivationRule</key>
        <dict>
            <key>NSExtensionActivationSupportsText</key>
            <true/>
            <key>NSExtensionActivationSupportsWebURLWithMaxCount</key>
            <integer>1</integer>
            <key>NSExtensionActivationSupportsWebPageWithMaxCount</key>
            <integer>1</integer>
        </dict>
        <key>NSExtensionJavaScriptPreprocessingFile</key>
        <string>Action</string>
    </dict>
    <key>NSExtensionPrincipalClass</key>
    <string>ShareViewController</string>
    <key>NSExtensionPointIdentifier</key>
    <string>com.apple.share-services</string>
</dict>
```

### 3.3 ShareViewController (SwiftUI)

```swift
import SwiftUI
import UniformTypeIdentifiers

@objc(ShareViewController)
class ShareViewController: UIViewController {
    override func viewDidLoad() {
        super.viewDidLoad()

        let contentView = UIHostingController(
            rootView: ShareView(extensionContext: self.extensionContext)
        )
        addChild(contentView)
        view.addSubview(contentView.view)

        contentView.view.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            contentView.view.topAnchor.constraint(equalTo: view.topAnchor),
            contentView.view.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            contentView.view.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            contentView.view.trailingAnchor.constraint(equalTo: view.trailingAnchor)
        ])
    }
}
```

### 3.4 获取分享内容

```swift
func extractSharedContent() async -> SharedContent? {
    guard let extensionItems = extensionContext?.inputItems as? [NSExtensionItem] else {
        return nil
    }

    for item in extensionItems {
        guard let attachments = item.attachments else { continue }

        for provider in attachments {
            // 尝试获取 URL
            if provider.hasItemConformingToTypeIdentifier(UTType.url.identifier) {
                if let url = try? await provider.loadItem(forTypeIdentifier: UTType.url.identifier) as? URL {
                    return SharedContent(url: url.absoluteString, title: nil)
                }
            }

            // 尝试获取文本 (可能包含链接)
            if provider.hasItemConformingToTypeIdentifier(UTType.text.identifier) {
                if let text = try? await provider.loadItem(forTypeIdentifier: UTType.text.identifier) as? String {
                    return SharedContent(url: text, title: nil)
                }
            }

            // 尝试获取 JavaScript 预处理结果 (网页标题)
            if provider.hasItemConformingToTypeIdentifier(UTType.propertyList.identifier) {
                if let dict = try? await provider.loadItem(forTypeIdentifier: UTType.propertyList.identifier) as? NSDictionary,
                   let jsResults = dict[NSExtensionJavaScriptPreprocessingResultsKey] as? NSDictionary {
                    let url = jsResults["URL"] as? String
                    let title = jsResults["title"] as? String
                    if let url = url {
                        return SharedContent(url: url, title: title)
                    }
                }
            }
        }
    }

    return nil
}
```

### 3.5 Action.js (获取网页标题)

```javascript
var Action = function() {};

Action.prototype = {
    run: function(parameters) {
        parameters.completionFunction({
            "URL": document.URL,
            "title": document.title
        });
    },

    finalize: function(parameters) {
        // 可选：页面操作
    }
};

var ExtensionPreprocessingJS = new Action();
```

## 4. 数据模型

### 4.1 SwiftData 模型

```swift
import SwiftData
import Foundation

@Model
final class Family {
    var id: String
    var pairingCode: String
    var createdAt: Date

    init(id: String, pairingCode: String) {
        self.id = id
        self.pairingCode = pairingCode
        self.createdAt = Date()
    }
}

@Model
final class Child {
    var id: String
    var familyId: String
    var name: String
    var avatarUrl: String?
    var dailyTimeLimitSec: Int
    var createdAt: Date

    init(id: String, familyId: String, name: String, dailyTimeLimitSec: Int = 3600) {
        self.id = id
        self.familyId = familyId
        self.name = name
        self.dailyTimeLimitSec = dailyTimeLimitSec
        self.createdAt = Date()
    }
}
```

### 4.2 App Groups 数据共享

```swift
import Foundation

class SharedStorage {
    static let shared = SharedStorage()

    private let defaults: UserDefaults? = {
        let groupIdentifier = "group.com.family.crab"
        return UserDefaults(suiteName: groupIdentifier)
    }()

    var jwtToken: String? {
        get { defaults?.string(forKey: "jwtToken") }
        set { defaults?.set(newValue, forKey: "jwtToken") }
    }

    var familyId: String? {
        get { defaults?.string(forKey: "familyId") }
        set { defaults?.set(newValue, forKey: "familyId") }
    }
}
```

## 5. 网络层

### 5.1 API Client

```swift
import Foundation

actor APIClient {
    static let shared = APIClient()

    private let baseURL = URL(string: "https://api.crab.family")!
    private let session = URLSession.shared

    func get<T: Decodable>(_ endpoint: String) async throws -> T {
        var request = URLRequest(url: baseURL.appendingPathComponent(endpoint))
        request.httpMethod = "GET"

        if let token = SharedStorage.shared.jwtToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            throw APIError.httpError(httpResponse.statusCode)
        }

        return try JSONDecoder().decode(T.self, from: data)
    }

    func post<T: Decodable, B: Encodable>(_ endpoint: String, body: B) async throws -> T {
        var request = URLRequest(url: baseURL.appendingPathComponent(endpoint))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = SharedStorage.shared.jwtToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        request.httpBody = try JSONEncoder().encode(body)

        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        guard (200...299).contains(httpResponse.statusCode) else {
            throw APIError.httpError(httpResponse.statusCode)
        }

        return try JSONDecoder().decode(T.self, from: data)
    }
}

enum APIError: Error {
    case invalidResponse
    case httpError(Int)
    case decodingError
}
```

## 6. 项目结构

```
ParentApp/
├── App/
│   ├── CrabApp.swift              # App 入口
│   └── ContentView.swift          # 主视图
├── Models/
│   ├── Family.swift               # 家庭模型
│   ├── Child.swift                # 孩子模型
│   └── APIModels.swift            # API 响应模型
├── Services/
│   ├── APIClient.swift            # 网络请求
│   └── SharedStorage.swift        # App Groups 存储
├── Views/
│   ├── OnboardingView.swift       # 首次启动/创建家庭
│   ├── PairingCodeView.swift      # 配对码展示
│   ├── ChildrenView.swift         # 孩子列表
│   └── SettingsView.swift         # 设置
└── ShareExtension/
    ├── ShareViewController.swift  # Extension 入口
    ├── ShareView.swift            # SwiftUI 分享页
    ├── Action.js                  # JS 预处理
    └── Info.plist                 # Extension 配置
```

## 7. 用户流程

### 7.1 首次使用

```
打开 App → 自动创建家庭 → 获取 JWT Token → 显示配对码 → 等待孩子配对
```

### 7.2 分享内容

```
外部 App (抖音/Safari/X) → 分享按钮 → 系统分享面板 → 选择 Crab
→ Extension 打开 → 解析链接 → 预览 → 选择孩子 → 发送 → 完成
```

## 8. 实施计划

| 阶段 | 任务 | 预估时间 |
|------|------|----------|
| Phase 1 | 项目初始化、App Groups、Share Extension Target | 0.5 天 |
| Phase 2 | SwiftData 模型、网络层、Token 管理 | 1 天 |
| Phase 3 | Share Extension UI、URL 解析、预览、发送 | 2 天 |
| Phase 4 | 主 App UI (配对码、孩子列表、设置) | 1.5 天 |
| **总计** | | **5 天** |

## 9. 后端 API 依赖

| API | 状态 | 用途 |
|-----|------|------|
| `POST /api/family/create` | ✅ 已实现 | 创建家庭 |
| `POST /api/family/pair` | ✅ 已实现 | 孩子配对 |
| `GET /api/children` | ✅ 已实现 | 获取孩子列表 |
| `POST /api/items/import` | ❌ 待实现 | 导入内容 |
| `POST /api/items/:id/assign` | ❌ 待实现 | 分配给孩子 |

## 10. 参考资料

- [iOS Share Extension with SwiftUI and SwiftData](https://www.merrell.dev/ios-share-extension-with-swiftui-and-swiftdata/)
- [How to build a basic Share Extension in Swift](https://tnvmadhav.me/guides/how-to-build-a-simple-share-extension-in-swift/)
- [Flutter iOS App Extensions 官方文档](https://docs.flutter.dev/platform-integration/ios/app-extensions)
- [Use Flutter UI inside iOS App Extensions](https://nemishah.com/blog/use-flutter-ui-inside-ios-app-extensions)
