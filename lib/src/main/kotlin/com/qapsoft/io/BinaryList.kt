package com.qapsoft.io

import java.io.ByteArrayInputStream
import java.io.ByteArrayOutputStream
import java.io.InputStream
import java.io.OutputStream
import java.io.StreamCorruptedException
import kotlin.math.max

open class BinaryList(protected val stream: BinaryStream, maxListSize:Int=1000) {
    open val schemaSize = 100
    open val indexSize = maxListSize * 4
    protected var realSize:Int = stream.getBytesAt(0, 4).asInt().let {
        if(it<0)
            0
        else
            it
    }

    val size get() = realSize

    fun add(writer:(out: OutputStream)->Unit):Int{
        val index = realSize
        val startPos = max(stream.length(), (schemaSize + indexSize).toLong())
        val os = ByteArrayOutputStream()
        writer(os)
        val bytes = os.toByteArray()
        stream.writeAt(startPos, bytes)
        stream.writeAt(schemaSize.toLong() + (index*4), startPos.toInt().toByteArray())
        realSize++
        stream.writeAt(0, realSize.toByteArray())
        return index
    }
    fun get(index:Int): InputStream{
        if(index < 0 || index >= size)
            throw IndexOutOfBoundsException()
        val startPos = (schemaSize + (index*4)).let { indexPos->
            val v = stream.getBytesAt(indexPos.toLong(),4).asInt()
            if(v<=0)
                throw StreamCorruptedException("Invalid start position for index $index")
            v
        }
        val endPos = (schemaSize + ((index+1)*4)).let {indexPos->
            if((index+1)>=realSize)
                stream.length().toInt()
            else{
                val v = stream.getBytesAt(indexPos.toLong(),4).asInt()
                if (v <= 0) stream.length().toInt() else v
            }
        }
        if (endPos < startPos) {
            throw StreamCorruptedException("end position ($endPos) < start position ($startPos)")
        }
        val bytes = stream.getBytesAt(startPos.toLong(), endPos-startPos)

        return ByteArrayInputStream(bytes)
    }
    fun toByteArray():ByteArray{
        return stream.getBytesAt(0, stream.length().toInt())
    }

}