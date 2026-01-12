package com.qapsoft.io

import kotlinx.serialization.SerialInfo

/**
 * Specifies the length of a column in bytes for String and RAW types in a BinaryTable.
 * This annotation is used when generating a BinaryTableHeader from a Serializable class.
 */
@SerialInfo
@Target(AnnotationTarget.PROPERTY)
annotation class BinaryColumnLength(val length: Int)
