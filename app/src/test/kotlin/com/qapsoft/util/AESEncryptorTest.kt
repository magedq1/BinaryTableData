package com.qapsoft.util

import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Test
import java.io.ByteArrayInputStream
import java.io.ByteArrayOutputStream

class AESEncryptorTest {

    @Test
    fun doCrypto() {
        val dataTest = EncryptionHelper.randomString(8192).toByteArray()

        val key = EncryptionHelper.randomString(32)
        val iv = AESEncryptor.generateIV()


        //encrypt
        var inputStream = ByteArrayInputStream(dataTest)
        var outputStream = ByteArrayOutputStream()

        AESEncryptor.encrypt(
            key = key,
            iv = iv,
            inputStream=inputStream,
            outputStream = outputStream
        )

        val encryptedData=outputStream.toByteArray()

        //decrypt
        inputStream = ByteArrayInputStream(encryptedData)
        outputStream= ByteArrayOutputStream()
        AESEncryptor.decrypt(
            key = key,
            iv = iv,
            inputStream=inputStream,
            outputStream = outputStream
        )

        //check
        assertArrayEquals(
            dataTest,
            outputStream.toByteArray()
        )

    }


}