package com.qapsoft.io

import org.junit.jupiter.api.Assertions.*

import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class CacheManagerTest {

    lateinit var cacheManager: CacheManager<String,String>
    @BeforeEach
    fun setUp() {
        cacheManager = CacheManager(50)
        for(i in 0..500){
            cacheManager.cache(i.toString(), i.toString())
            if(i>101)
                cacheManager.getOrNull("101")
        }
    }



    @Test
    fun getOrNull() {
        assertNull(cacheManager.getOrNull("100"))
        assertNotNull(cacheManager.getOrNull("101"))
        assertNotNull(cacheManager.getOrNull("470"))
    }
}