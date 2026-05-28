package com.family.crab.child.data.api

import com.family.crab.child.data.model.PairingResponse
import com.family.crab.child.data.model.ContentListResponse
import retrofit2.Response
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query

interface ApiService {
    
    /**
     * Pair device with parent using pairing code
     */
    @POST("api/v1/device/pair")
    suspend fun pairDevice(@Query("code") pairingCode: String): Response<PairingResponse>
    
    /**
     * Get allowed content list for the child
     */
    @GET("api/v1/content")
    suspend fun getContentList(): Response<ContentListResponse>
    
    /**
     * Report playback progress to parent
     */
    @POST("api/v1/content/{contentId}/progress")
    suspend fun reportProgress(
        @Path("contentId") contentId: String,
        @Query("position") positionMs: Long,
        @Query("duration") durationMs: Long
    ): Response<Unit>
}
