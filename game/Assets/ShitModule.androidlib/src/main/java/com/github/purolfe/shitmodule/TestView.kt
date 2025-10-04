package com.github.purolfe.shitmodule

import androidx.compose.runtime.Composable
import org.mozilla.geckoview.GeckoSession

class TestView {
    init {
        GeckoSession.PRIORITY_HIGH
    }

    @Composable
    fun TestUI() {

    }
}