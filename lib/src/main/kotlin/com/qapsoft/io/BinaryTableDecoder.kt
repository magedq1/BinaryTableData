package com.qapsoft.io

import kotlinx.serialization.DeserializationStrategy
import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.serializer
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.CompositeDecoder
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.modules.EmptySerializersModule
import kotlinx.serialization.modules.SerializersModule

@OptIn(ExperimentalSerializationApi::class)
class BinaryTableDecoder(
    private val table: BinaryTableData,
    private val rowIndex: Int
) : Decoder, CompositeDecoder {
    private var elementIndex = 0
    override val serializersModule: SerializersModule = EmptySerializersModule()

    override fun beginStructure(descriptor: SerialDescriptor): CompositeDecoder {
        return this
    }

    override fun endStructure(descriptor: SerialDescriptor) {
    }

    override fun decodeElementIndex(descriptor: SerialDescriptor): Int {
        while (elementIndex < descriptor.elementsCount) {
            val name = descriptor.getElementName(elementIndex)
            if (table.header.columnsInfoByName.containsKey(name)) {
                return elementIndex++
            }
            elementIndex++
        }
        return CompositeDecoder.DECODE_DONE
    }

    private fun getColumnName(descriptor: SerialDescriptor, index: Int): String {
        return descriptor.getElementName(index)
    }

    override fun decodeBooleanElement(descriptor: SerialDescriptor, index: Int): Boolean {
        return decodeIntElement(descriptor, index) != 0
    }

    override fun decodeByteElement(descriptor: SerialDescriptor, index: Int): Byte {
        return decodeIntElement(descriptor, index).toByte()
    }

    override fun decodeCharElement(descriptor: SerialDescriptor, index: Int): Char {
        return decodeIntElement(descriptor, index).toChar()
    }
    
    override fun decodeDoubleElement(descriptor: SerialDescriptor, index: Int): Double {
        return table.getDouble(rowIndex, getColumnName(descriptor, index))
    }
    
    override fun decodeFloatElement(descriptor: SerialDescriptor, index: Int): Float {
        return decodeDoubleElement(descriptor, index).toFloat()
    }
    
    override fun decodeIntElement(descriptor: SerialDescriptor, index: Int): Int {
        return table.getInt(rowIndex, getColumnName(descriptor, index))
    }
    
    override fun decodeLongElement(descriptor: SerialDescriptor, index: Int): Long {
        return table.getLong(rowIndex, getColumnName(descriptor, index))
    }
    
    override fun decodeStringElement(descriptor: SerialDescriptor, index: Int): String {
        return table.getString(rowIndex, getColumnName(descriptor, index))
    }

    override fun <T> decodeSerializableElement(
        descriptor: SerialDescriptor,
        index: Int,
        deserializer: DeserializationStrategy<T>,
        previousValue: T?
    ): T {
        val name = descriptor.getElementName(index)
        val columnInfo = table.header.columnsInfoByName[name]
        
        if (columnInfo?.binaryColumn?.columnType == BinaryColumn.ColumnType.RAW) {
            val serialName = deserializer.descriptor.serialName
            if (serialName == "kotlin.ByteArray") {
                @Suppress("UNCHECKED_CAST")
                return table.getValue(rowIndex, name) as T
            }
            if (serialName == "com.qapsoft.io.BinaryStream" || serialName == "com.qapsoft.io.BinaryStreamReader") {
                @Suppress("UNCHECKED_CAST")
                return table.getValueAsStream(rowIndex, name) as T
            }
        }
        return deserializer.deserialize(this)
    }

    override fun decodeShortElement(
        descriptor: SerialDescriptor,
        index: Int
    ): Short {
        return decodeIntElement(descriptor, index).toShort()
    }

    override fun <T : Any> decodeNullableSerializableElement(
        descriptor: SerialDescriptor,
        index: Int,
        deserializer: DeserializationStrategy<T?>,
        previousValue: T?
    ): T? {
         return if (table.header.columnsInfoByName.containsKey(descriptor.getElementName(index))) {
             deserializer.deserialize(this)
         } else {
             null
         }
    }

    override fun decodeInlineElement(descriptor: SerialDescriptor, index: Int): Decoder = this

    // Unsupported direct decode methods (since we expect structure decoding via index)
    override fun decodeBoolean(): Boolean = throw UnsupportedOperationException("Direct decodeBoolean not supported")
    override fun decodeByte(): Byte = throw UnsupportedOperationException("Direct decodeByte not supported")
    override fun decodeShort(): Short = throw UnsupportedOperationException("Direct decodeShort not supported")
    override fun decodeChar(): Char = throw UnsupportedOperationException("Direct decodeChar not supported")
    override fun decodeEnum(enumDescriptor: SerialDescriptor): Int = throw UnsupportedOperationException("Direct decodeEnum not supported")
    override fun decodeFloat(): Float = throw UnsupportedOperationException("Direct decodeFloat not supported")
    override fun decodeInline(descriptor: SerialDescriptor): Decoder {
        return this
    }

    override fun decodeInt(): Int = throw UnsupportedOperationException("Direct decodeInt not supported")
    override fun decodeLong(): Long = throw UnsupportedOperationException("Direct decodeLong not supported")
    override fun decodeString(): String = throw UnsupportedOperationException("Direct decodeString not supported")
    override fun decodeDouble(): Double = throw UnsupportedOperationException("Direct decodeDouble not supported")
    override fun decodeNull(): Nothing? = null
    override fun decodeNotNullMark(): Boolean = true
}

inline fun <reified T> BinaryTableData.decodeRow(rowIndex: Int): T {
    val decoder = BinaryTableDecoder(this, rowIndex)
    return decoder.decodeSerializableValue(serializer<T>())
}

inline fun <reified T> BinaryTableData.decodeAll(): Sequence<T> {
    return sequence {
        for (i in 0 until header.maxRowsCount) {
             yield(decodeRow<T>(i))
        }
    }
}
