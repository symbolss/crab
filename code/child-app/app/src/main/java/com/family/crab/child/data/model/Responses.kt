package com.family.crab.child.data.model

import com.google.gson.annotations.SerializedName

data class PairingResponse(
    @SerializedName("success")
    val success: Boolean,
    
    @SerializedName("token")
    val token: String?,
    
    @SerializedName("child_name")
    val childName: String?,
    
    @SerializedName("error")
    val error: String?
)

data class ContentListResponse(
    @SerializedName("contents")
    val contents: List<ContentItem>
)

data class ContentItem(
    @SerializedName("id")
    val id: String,
    
    @SerializedName("title")
    val title: String,
    
    @SerializedName("thumbnail_url")
    val thumbnailUrl: String?,
    
    @SerializedName("video_url")
    val videoUrl: String,
    
    @SerializedName("duration_ms")
    val durationMs: Long,
    
    @SerializedName("category")
    val category: String?,
    
    @SerializedName("description")
    val description: String?
)
