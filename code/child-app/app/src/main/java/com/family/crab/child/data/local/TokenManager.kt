package com.family.crab.child.data.local

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenManager @Inject constructor(
    private val dataStore: DataStore<Preferences>
) {
    companion object {
        private val TOKEN_KEY = stringPreferencesKey("auth_token")
        private val CHILD_NAME_KEY = stringPreferencesKey("child_name")
        private val DEVICE_ID_KEY = stringPreferencesKey("device_id")
    }

    suspend fun saveToken(token: String) {
        dataStore.edit { prefs ->
            prefs[TOKEN_KEY] = token
        }
    }

    suspend fun getToken(): String? {
        return dataStore.data.map { prefs ->
            prefs[TOKEN_KEY]
        }.first()
    }

    suspend fun saveChildName(name: String) {
        dataStore.edit { prefs ->
            prefs[CHILD_NAME_KEY] = name
        }
    }

    suspend fun getChildName(): String? {
        return dataStore.data.map { prefs ->
            prefs[CHILD_NAME_KEY]
        }.first()
    }

    suspend fun saveDeviceId(deviceId: String) {
        dataStore.edit { prefs ->
            prefs[DEVICE_ID_KEY] = deviceId
        }
    }

    suspend fun getDeviceId(): String? {
        return dataStore.data.map { prefs ->
            prefs[DEVICE_ID_KEY]
        }.first()
    }

    suspend fun clearAll() {
        dataStore.edit { prefs ->
            prefs.clear()
        }
    }

    fun hasTokenFlow() = dataStore.data.map { prefs ->
        prefs[TOKEN_KEY] != null
    }
}
