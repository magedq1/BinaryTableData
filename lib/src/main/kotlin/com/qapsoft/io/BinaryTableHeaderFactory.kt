package com.qapsoft.io

import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.KSerializer
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.SerialDescriptor

object BinaryTableHeaderFactory {
    @OptIn(ExperimentalSerializationApi::class)
    fun <T> createHeaderFrom(serializer: KSerializer<T>, maxRows: Int): BinaryTableHeader {
        val descriptor = serializer.descriptor
        val columns = mutableListOf<BinaryColumn>()

        for (i in 0 until descriptor.elementsCount) {
            val name = descriptor.getElementName(i)
            val kind = descriptor.getElementDescriptor(i).kind
            val annotations = descriptor.getElementAnnotations(i)
            
            val lengthAnnotation = annotations.filterIsInstance<BinaryColumnLength>().firstOrNull()
            
            val column = when (kind) {
                PrimitiveKind.INT -> BinaryColumn.Int(name)
                PrimitiveKind.LONG -> BinaryColumn.Long(name)
                PrimitiveKind.DOUBLE -> BinaryColumn.Double(name)
                PrimitiveKind.STRING -> {
                    val length = lengthAnnotation?.length ?: 64 // Default to 64 if not specified
                    BinaryColumn.String(name, length)
                }
                else -> {
                    // Check for RAW types based on serialName or if it's explicitly annotated/handled as RAW provided in previous steps
                    val serialName = descriptor.getElementDescriptor(i).serialName
                    if (serialName == "kotlin.ByteArray" || 
                        serialName == "com.qapsoft.io.BinaryStream" || 
                        serialName == "com.qapsoft.io.BinaryStreamReader") {
                        val length = lengthAnnotation?.length ?: throw IllegalArgumentException("Field '$name' of type '$serialName' (RAW) must be annotated with @BinaryColumnLength")
                        BinaryColumn.Raw(name, length)
                    } else {
                         // Default fallback or error?
                         // Depending on requirement, we might treat unknown as error or try string.
                         // For now, let's treat other primitives as error or strict mapping.
                         throw IllegalArgumentException("Unsupported column type for field '$name': $kind / $serialName")
                    }
                }
            }
            columns.add(column)
        }

        return BinaryTableHeader(
            maxRowsCount = maxRows,
            columns = columns
        )
    }
}
