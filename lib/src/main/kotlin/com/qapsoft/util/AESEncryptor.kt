package com.qapsoft.util

import java.io.File
import java.io.FileInputStream
import java.io.FileNotFoundException
import java.io.FileOutputStream
import java.io.InputStream
import java.io.OutputStream
import java.security.Key
import java.security.MessageDigest
import java.security.NoSuchAlgorithmException
import java.security.SecureRandom
import javax.crypto.Cipher
import javax.crypto.CipherInputStream
import javax.crypto.CipherOutputStream
import javax.crypto.spec.IvParameterSpec
import javax.crypto.spec.SecretKeySpec
import kotlin.math.min



object AESEncryptor {
    @Throws(Exception::class)
    fun encrypt(key: String, iv: ByteArray, inputStream: InputStream, outputStream: OutputStream) {
        doCrypto(Cipher.ENCRYPT_MODE, key, iv, inputStream, outputStream)
    }

    @Throws(Exception::class)
    fun decrypt(key: String, iv: ByteArray, inputStream: InputStream, outputStream: OutputStream) {
        doCrypto(Cipher.DECRYPT_MODE, key, iv, inputStream, outputStream)
    }

    @Throws(FileNotFoundException::class, Exception::class)
    fun doCrypto(decryptMode: Int, key: String, iv: ByteArray, inputFile: File, outputFile: File) {
        doCrypto(
            decryptMode,
            key,
            iv,
            (FileInputStream(inputFile)) as InputStream,
            (FileOutputStream(outputFile)) as OutputStream
        )
    }

    @Throws(Exception::class)
    fun doCrypto(
        cipherMode: Int,
        key: String,
        iv: ByteArray,
        inputStream: InputStream,
        outputStream: OutputStream
    ) {
        val ALGORITHM = "AES/CBC/PKCS5Padding"
        val cipher = Cipher.getInstance(ALGORITHM)
        val secretKey: Key = SecretKeySpec(convertPasswordTo32Bytes(key), "AES")
        cipher.init(cipherMode, secretKey, IvParameterSpec(iv))
        val buffer = ByteArray(1024)
        var bytesRead: Int
        when (cipherMode) {
            1 -> {
                val coutputStream = CipherOutputStream(outputStream, cipher)

                while ((inputStream.read(buffer).also { bytesRead = it }) != -1) {
                    coutputStream.write(buffer, 0, bytesRead)
                }

                coutputStream.close()
            }

            2 -> {
                val cipherInputStream = CipherInputStream(inputStream, cipher)

                while ((cipherInputStream.read(buffer).also { bytesRead = it }) != -1) {
                    outputStream.write(buffer, 0, bytesRead)
                }

                cipherInputStream.close()
                outputStream.flush()
                outputStream.close()
            }
        }

        inputStream.close()
    }

    @Throws(NoSuchAlgorithmException::class)
    fun convertPasswordTo32Bytes(password: String): ByteArray {
        val digest = MessageDigest.getInstance("SHA-256")
        val passwordBytes = password.toByteArray()
        digest.update(passwordBytes)
        val hash = digest.digest()
        val paddedHash = ByteArray(32)
        System.arraycopy(
            hash, 0, paddedHash, 0,
            min(hash.size.toDouble(), 32.0).toInt()
        )
        return paddedHash
    }

    @Throws(NoSuchAlgorithmException::class)
    fun generateIV(): ByteArray {
        val iv = ByteArray(16)
        SecureRandom.getInstanceStrong().nextBytes(iv)
        return iv
    }
}