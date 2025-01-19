package com.qapsoft.io

class CacheManager<K,T>(val capacity:Int=50) {
    private val cache = object : LinkedHashMap<K, T>(capacity, 0.75f, true) {
        override fun removeEldestEntry(eldest: MutableMap.MutableEntry<K, T>): Boolean {
            return size > capacity
        }
    }

    fun cache(key: K, value: T): T = value.also { cache[key] = value }

    fun clear() = cache.clear()

    fun containsKey(key: K) = key in cache

    fun remove(key: K) {
        cache.remove(key)
    }

    operator fun get(key: K):T?{
        return getOrNull(key)
    }
    operator fun set(key: K, value: T){
        cache(key, value)
    }

    fun getOrNull(key: K): T? = cache[key]

    val size: Int get() = cache.size
}