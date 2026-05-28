import SwiftUI

@main
struct CrabApp: App {
    @StateObject private var appState = AppState()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(appState)
                .onAppear {
                    appState.initialize()
                }
        }
    }
}

// MARK: - App State

@MainActor
class AppState: ObservableObject {
    @Published var isInitialized = false
    @Published var familyId: String?
    @Published var pairingCode: String?
    @Published var hasFamily: Bool = false

    private let storage = SharedStorage.shared
    private let apiClient = APIClient.shared

    func initialize() {
        // Check if we already have a family
        if let savedFamilyId = storage.familyId,
           let _ = storage.jwtToken {
            self.familyId = savedFamilyId
            self.hasFamily = true
        }

        self.isInitialized = true
    }

    func createFamily() async throws {
        let response = try await apiClient.post("api/family/create", body: EmptyRequest())

        self.familyId = response.familyId
        self.pairingCode = response.pairingCode
        self.hasFamily = true

        // Save to shared storage
        storage.familyId = response.familyId
        storage.jwtToken = response.parentToken
    }
}

// MARK: - Empty Request

struct EmptyRequest: Encodable {}
