package com.qapsoft.io

import java.io.ByteArrayOutputStream
import java.util.zip.Deflater
import java.util.zip.Inflater




object DataCompression {

    fun deflateData(data: ByteArray, start:Int,offset:Int): ByteArray {
        val deflater = Deflater()
        deflater.setInput(data, start, offset)
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
    fun deflateData(data: ByteArray): ByteArray {
        return deflateData(data, 0 , data.size)
    }

    fun inflateData(data: ByteArray): ByteArray? {
        val inflater = Inflater()
        inflater.setInput(data)

        val outputStream = ByteArrayOutputStream()
        val buffer = ByteArray(1024)
        try {
            while (!inflater.finished()) {
                val inflatedSize = inflater.inflate(buffer)
                if(inflatedSize<=0)
                    break
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

}