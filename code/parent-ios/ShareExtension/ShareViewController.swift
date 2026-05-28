import UIKit
import Social
import SwiftUI

class ShareViewController: SLComposeServiceViewController {

    override func isContentValid() -> Bool {
        // 检查是否有家庭 ID
        return SharedStorage.shared.familyId != nil
    }

    override func viewDidLoad() {
        super.viewDidLoad()

        // 使用 SwiftUI 视图
        let shareView = ShareView()
        let hostingController = UIHostingController(rootView: shareView)

        // 移除默认视图，添加 SwiftUI 视图
        view.subviews.forEach { $0.removeFromSuperview() }

        addChild(hostingController)
        view.addSubview(hostingController.view)
        hostingController.view.translatesAutoresizingMaskIntoConstraints = false
        NSLayoutConstraint.activate([
            hostingController.view.topAnchor.constraint(equalTo: view.topAnchor),
            hostingController.view.bottomAnchor.constraint(equalTo: view.bottomAnchor),
            hostingController.view.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            hostingController.view.trailingAnchor.constraint(equalTo: view.trailingAnchor)
        ])
        hostingController.didMove(toParent: self)
    }

    override func didSelectPost() {
        // 由 SwiftUI 视图处理
    }

    override func configurationItems() -> [Any]! {
        return []
    }
}
