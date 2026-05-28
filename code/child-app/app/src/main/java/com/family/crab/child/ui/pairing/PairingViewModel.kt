package com.family.crab.child.ui.pairing

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.family.crab.child.data.api.ApiService
import com.family.crab.child.data.local.TokenManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PairingViewModel @Inject constructor(
    private val apiService: ApiService,
    private val tokenManager: TokenManager
) : ViewModel() {
    
    private val _pairingState = MutableLiveData<PairingState>()
    val pairingState: LiveData<PairingState> = _pairingState
    
    sealed class PairingState {
        object Idle : PairingState()
        object Loading : PairingState()
        data class Success(val childName: String) : PairingState()
        data class Error(val message: String) : PairingState()
    }
    
    fun pairDevice(pairingCode: String) {
        if (pairingCode.isBlank()) {
            _pairingState.value = PairingState.Error("Please enter a pairing code")
            return
        }
        
        viewModelScope.launch {
            _pairingState.value = PairingState.Loading
            
            try {
                val response = apiService.pairDevice(pairingCode.trim())
                
                if (response.isSuccessful && response.body()?.success == true) {
                    val body = response.body()!!
                    tokenManager.saveToken(body.token!!)
                    body.childName?.let { tokenManager.saveChildName(it) }
                    _pairingState.value = PairingState.Success(body.childName ?: "Child")
                } else {
                    val errorMsg = response.body()?.error ?: "Pairing failed"
                    _pairingState.value = PairingState.Error(errorMsg)
                }
            } catch (e: Exception) {
                _pairingState.value = PairingState.Error(
                    e.message ?: "Network error occurred"
                )
            }
        }
    }
    
    fun resetState() {
        _pairingState.value = PairingState.Idle
    }
}
