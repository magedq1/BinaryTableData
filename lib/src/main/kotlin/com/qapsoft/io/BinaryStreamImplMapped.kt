package com.qapsoft.io

import java.io.Closeable
import java.io.File
import java.io.IOException
import java.io.RandomAccessFile
import java.nio.MappedByteBuffer
import java.nio.channels.FileChannel

open class BinaryStreamImplMapped(file: File, private val size:Int, val mode: Mode=Mode.READ_WRITE):BinaryStream, Closeable{
    enum class Mode{
        READ, READ_WRITE
    }
    val channel: FileChannel
    val mappedByteBuffer: MappedByteBuffer
    init {
        channel = RandomAccessFile(file, when(mode){
            Mode.READ -> "r"
            Mode.READ_WRITE -> "rw"
        }).channel
        mappedByteBuffer = channel.map(when(mode){
            Mode.READ -> FileChannel.MapMode.READ_ONLY
            Mode.READ_WRITE -> FileChannel.MapMode.READ_WRITE
        }, 0, size.toLong())
        channel.read(mappedByteBuffer)
    }
    override fun readAt(pos: Long,buffer:ByteArray, start: Int, offset: Int): Int {
        this.mappedByteBuffer.position(pos.toInt())
        this.mappedByteBuffer.get(buffer, start, offset)
        return offset
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos,buffer,0 , buffer.size)
    }

    override fun length(): Long {
        return size.toLong()
    }

    override fun writeAt(pos: Long, buffer:ByteArray,start: Int, offset: Int) {
        if(mode==Mode.READ)
            throw IOException("this stream can read only!")
        this.mappedByteBuffer.position(pos.toInt())
        this.mappedByteBuffer.put(buffer, start, offset)
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos,buffer, 0 ,buffer.size)
    }

    override fun close() {
        mappedByteBuffer.force()
        channel.close()
    }
}