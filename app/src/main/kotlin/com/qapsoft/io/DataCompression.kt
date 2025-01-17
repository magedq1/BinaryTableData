package com.qapsoft.io

import java.io.ByteArrayOutputStream
import java.util.zip.DataFormatException
import java.util.zip.Deflater
import java.util.zip.Inflater




object DataCompression {

    fun deflateData(data: ByteArray): ByteArray {
        val deflater = Deflater()
        deflater.setInput(data)
        deflater.finish()

        val outputStream = ByteArrayOutputStream()
        val buffer = ByteArray(1024)
        while (!deflater.finished()) {
            val compressedSize = deflater.deflate(buffer)
            outputStream.write(buffer, 0, compressedSize)
        }
        deflater.end()
        return outputStream.toByteArray()
    }

    fun inflateData(data: ByteArray): ByteArray? {
        val inflater = Inflater()
        inflater.setInput(data)

        val outputStream = ByteArrayOutputStream()
        val buffer = ByteArray(1024)
        try {
            while (!inflater.finished()) {
                val inflatedSize = inflater.inflate(buffer)
                outputStream.write(buffer, 0, inflatedSize)
            }
        } catch (e: Exception) {
            // Handle exception (e.g., incorrect data format)
            e.printStackTrace()
            return ByteArray(0) // Or throw the exception
        } finally {
            inflater.end()
        }
        return outputStream.toByteArray()
    }

    /*fun inflateData(data: ByteArray): ByteArray? {
        val bos = ByteArrayOutputStream()
        inflater(data, bos)
        val res = bos.toByteArray()
        bos.close()
        return res
    }*/

    private fun inflater(data: ByteArray, bos: ByteArrayOutputStream) {
        val inflater = Inflater()
        inflater.setInput(data)
        val arrayOfByte = ByteArray(8192)
        while (!inflater.finished()) {
            val i = inflater.inflate(arrayOfByte)
            bos.write(arrayOfByte, 0, i)
        }
        inflater.end()
    }
}