package com.qapsoft.io

import java.io.Closeable
import java.io.File
import java.io.IOException
import java.io.RandomAccessFile
import java.nio.MappedByteBuffer
import java.nio.channels.FileChannel

open class BinaryStreamImpRandomAccess(file: File, val mode: Mode=Mode.READ_WRITE):BinaryStream, Closeable{
    enum class Mode{
        READ, READ_WRITE
    }
    val rFile: RandomAccessFile
    init {
        rFile = RandomAccessFile(file, when(mode){
            Mode.READ -> "r"
            Mode.READ_WRITE -> "rw"
        })

    }
    override fun readAt(pos: Long,buffer:ByteArray, start: Int, offset: Int): Int {
        // Use FileChannel for thread-safe positional reads
        return rFile.channel.read(java.nio.ByteBuffer.wrap(buffer, start, offset), pos)
    }

    override fun readAt(pos: Long, buffer: ByteArray): Int {
        return readAt(pos,buffer,0 , buffer.size)
    }

    override fun length(): Long {
        return rFile.length()
    }

    override fun writeAt(pos: Long, buffer:ByteArray,start: Int, offset: Int) {
        if(mode==Mode.READ)
            throw IOException("this stream can read only!")
        // Use FileChannel for thread-safe positional writes
        rFile.channel.write(java.nio.ByteBuffer.wrap(buffer, start, offset), pos)
    }

    override fun writeAt(pos: Long, buffer: ByteArray) {
        writeAt(pos,buffer, 0 ,buffer.size)
    }

    override fun close() {
        rFile.close()
    }
}