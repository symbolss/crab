package com.family.crab.child.ui.content

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.family.crab.child.data.api.ApiService
import com.family.crab.child.data.model.ContentItem
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ContentViewModel @Inject constructor(
    private val apiService: ApiService
) : ViewModel() {
    
    private val _contentList = MutableLiveData<List<ContentItem>>()
    val contentList: LiveData<List<ContentItem>> = _contentList
    
    private val _isLoading = MutableLiveData<Boolean>()
    val isLoading: LiveData<Boolean> = _isLoading
    
    private val _error = MutableLiveData<String?>()
    val error: LiveData<String?> = _error
    
    init {
        loadContent()
    }
    
    fun loadContent() {
        viewModelScope.launch {
            _isLoading.value = true
            _error.value = null
            
            try {
                val response = apiService.getContentList()
                if (response.isSuccessful && response.body() != null) {
                    _contentList.value = response.body()!!.contents
                } else {
                    _error.value = "Failed to load content"
                }
            } catch (e: Exception) {
                _error.value = e.message ?: "Network error"
            } finally {
                _isLoading.value = false
            }
        }
    }
}
