package com.qapsoft.util

import java.security.SecureRandom
import java.util.Random

object EncryptionHelper {
    private val CHARACTERS: String =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+"


    fun randomString(length: Int): String {
        val random: Random = SecureRandom()
        val passwordBuilder = StringBuilder(length)

        for (i in 0 until length) {
            val index =
                random.nextInt(CHARACTERS.length)
            passwordBuilder.append(CHARACTERS[index])
        }

        return passwordBuilder.toString()
    }

}