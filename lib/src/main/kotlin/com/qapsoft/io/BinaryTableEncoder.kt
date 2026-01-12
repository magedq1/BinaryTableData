package com.qapsoft.io

import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.SerializationStrategy
import kotlinx.serialization.serializer
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.CompositeEncoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.modules.EmptySerializersModule
import kotlinx.serialization.modules.SerializersModule

@OptIn(ExperimentalSerializationApi::class)
class BinaryTableEncoder(
    private val table: BinaryTableData,
    private val rowIndex: Int
) : Encoder, CompositeEncoder {

    override val serializersModule: SerializersModule = EmptySerializersModule()

    override fun beginStructure(descriptor: SerialDescriptor): CompositeEncoder {
        return this
    }

    override fun endStructure(descriptor: SerialDescriptor) {
    }

    private fun getColumnName(descriptor: SerialDescriptor, index: Int): String {
        return descriptor.getElementName(index)
    }

    override fun encodeBooleanElement(descriptor: SerialDescriptor, index: Int, value: Boolean) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        val intValue = if (value) 1 else 0
        table.setValue(rowIndex, name, intValue)
    }

    override fun encodeByteElement(descriptor: SerialDescriptor, index: Int, value: Byte) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value.toInt())
    }

    override fun encodeCharElement(descriptor: SerialDescriptor, index: Int, value: Char) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value.code)
    }

    override fun encodeDoubleElement(descriptor: SerialDescriptor, index: Int, value: Double) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value)
    }

    override fun encodeFloatElement(descriptor: SerialDescriptor, index: Int, value: Float) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value.toDouble())
    }

    override fun encodeIntElement(descriptor: SerialDescriptor, index: Int, value: Int) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value)
    }

    override fun encodeLongElement(descriptor: SerialDescriptor, index: Int, value: Long) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value)
    }

    override fun encodeShortElement(descriptor: SerialDescriptor, index: Int, value: Short) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value.toInt())
    }

    override fun encodeStringElement(descriptor: SerialDescriptor, index: Int, value: String) {
        val name = getColumnName(descriptor, index)
        if (!table.header.columnsInfoByName.containsKey(name)) return
        table.setValue(rowIndex, name, value)
    }

    override fun <T> encodeSerializableElement(
        descriptor: SerialDescriptor,
        index: Int,
        serializer: SerializationStrategy<T>,
        value: T
    ) {
        val name = descriptor.getElementName(index)
        val columnInfo = table.header.columnsInfoByName[name] ?: return

        if (columnInfo?.binaryColumn?.columnType == BinaryColumn.ColumnType.RAW) {
            val serialName = serializer.descriptor.serialName
            if (serialName == "kotlin.ByteArray") {
                table.setValue(rowIndex, name, value as ByteArray)
                return
            }
            if (serialName == "com.qapsoft.io.BinaryStream") {
                 val stream = value as BinaryStream
                 val bytes = stream.getBytesAt(0, stream.length().toInt())
                 table.setValue(rowIndex, name, bytes)
                 return
            }
            if (serialName == "com.qapsoft.io.BinaryStreamReader") {
                val reader = value as BinaryStreamReader
                val bytes = reader.getBytesAt(0, reader.length().toInt())
                table.setValue(rowIndex, name, bytes)
                return
            }
        }
        serializer.serialize(this, value)
    }

    override fun <T : Any> encodeNullableSerializableElement(
        descriptor: SerialDescriptor,
        index: Int,
        serializer: SerializationStrategy<T>,
        value: T?
    ) {
        if (value != null && table.header.columnsInfoByName.containsKey(descriptor.getElementName(index))) {
            encodeSerializableElement(descriptor, index, serializer, value)
        }
    }

    override fun encodeInlineElement(descriptor: SerialDescriptor, index: Int): Encoder = this

    // Unsupported direct encode methods
    override fun encodeBoolean(value: Boolean) = throw UnsupportedOperationException("Direct encodeBoolean not supported")
    override fun encodeByte(value: Byte) = throw UnsupportedOperationException("Direct encodeByte not supported")
    override fun encodeChar(value: Char) = throw UnsupportedOperationException("Direct encodeChar not supported")
    override fun encodeDouble(value: Double) = throw UnsupportedOperationException("Direct encodeDouble not supported")
    override fun encodeEnum(enumDescriptor: SerialDescriptor, index: Int) = throw UnsupportedOperationException("Direct encodeEnum not supported")
    override fun encodeFloat(value: Float) = throw UnsupportedOperationException("Direct encodeFloat not supported")
    override fun encodeInline(descriptor: SerialDescriptor): Encoder = this
    override fun encodeInt(value: Int) = throw UnsupportedOperationException("Direct encodeInt not supported")
    override fun encodeLong(value: Long) = throw UnsupportedOperationException("Direct encodeLong not supported")
    override fun encodeNull() {} // Ignore nulls?
    override fun encodeShort(value: Short) = throw UnsupportedOperationException("Direct encodeShort not supported")
    override fun encodeString(value: String) = throw UnsupportedOperationException("Direct encodeString not supported")
}

inline fun <reified T> BinaryTableData.encodeRow(rowIndex: Int, value: T) {
    val encoder = BinaryTableEncoder(this, rowIndex)
    encoder.encodeSerializableValue(serializer<T>(), value)
}
