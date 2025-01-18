package com.qapsoft.util

import java.io.ByteArrayOutputStream
import java.io.IOException
import java.io.InputStream
import java.security.Key
import java.security.KeyFactory
import java.security.KeyPair
import java.security.KeyPairGenerator
import java.security.NoSuchAlgorithmException
import java.security.PrivateKey
import java.security.PublicKey
import java.security.Signature
import java.security.interfaces.RSAKey
import java.security.interfaces.RSAPrivateKey
import java.security.interfaces.RSAPublicKey
import java.security.spec.InvalidKeySpecException
import java.security.spec.PKCS8EncodedKeySpec
import java.security.spec.X509EncodedKeySpec
import javax.crypto.Cipher


object RSAEncryptHelper {

    @Throws(Exception::class)
    fun input2bytes(inputStream: InputStream): ByteArray {
        val os = ByteArrayOutputStream()
        val buff = ByteArray(9000)

        var z: Int
        while ((inputStream.read(buff).also { z = it }) > 0) {
            os.write(buff, 0, z)
        }

        return os.toByteArray()
    }
    @Throws(NoSuchAlgorithmException::class)
    fun GenerateKeyPair(keysize: Int): KeyPair {
        val keyPairGen = KeyPairGenerator.getInstance("RSA")
        keyPairGen.initialize(keysize)
        val pair = keyPairGen.generateKeyPair()
        return pair
    }

    @Throws(
        InvalidKeySpecException::class,
        NoSuchAlgorithmException::class,
        IOException::class
    )
    fun readPublicKey(inputStream: InputStream): RSAPublicKey {
        val publicKeySpec = X509EncodedKeySpec(input2bytes(inputStream))
        val importedPublicKey =
            KeyFactory.getInstance("RSA").generatePublic(publicKeySpec) as RSAPublicKey
        return importedPublicKey
    }

    @Throws(
        InvalidKeySpecException::class,
        NoSuchAlgorithmException::class,
        IOException::class
    )

    fun readPrivateKey(inputStream: InputStream): RSAPrivateKey {
        val spec = PKCS8EncodedKeySpec(input2bytes(inputStream))
        val kf = KeyFactory.getInstance("RSA")
        val privKey = kf.generatePrivate(spec) as RSAPrivateKey
        return privKey
    }

    @Throws(Exception::class)
    fun processData(key: RSAKey, encryptMode: Boolean, input: ByteArray): ByteArray {
        val cipher = Cipher.getInstance("RSA")
        cipher.init(if (encryptMode) 1 else 2, key as Key?)
        cipher.update(input)
        return cipher.doFinal()
    }

    @Throws(Exception::class)
    fun sign(data: ByteArray, privateKey: PrivateKey?): ByteArray {
        val privateSignature = Signature.getInstance("SHA256withRSA")
        privateSignature.initSign(privateKey)
        privateSignature.update(data)
        val signature = privateSignature.sign()
        return signature
    }

    @Throws(Exception::class)
    fun verify(data: ByteArray, signature: ByteArray, publicKey: PublicKey): Boolean {
        val publicSignature = Signature.getInstance("SHA256withRSA")
        publicSignature.initVerify(publicKey)
        publicSignature.update(data)
        return publicSignature.verify(signature)
    }
}