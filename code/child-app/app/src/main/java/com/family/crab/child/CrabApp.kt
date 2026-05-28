package com.family.crab.child

import android.app.Application
import dagger.hilt.android.HiltAndroidApp

@HiltAndroidApp
class CrabApp : Application() {
    
    override fun onCreate() {
        super.onCreate()
        // Application initialization
    }
}
