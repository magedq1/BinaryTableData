package com.qapsoft.io


data class BinaryColumn (
    val name:String,
    val length:Int,
    val columnType: ColumnType = ColumnType.RAW
){
    companion object{
        fun Raw (name: String, length: Int):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = length,
                columnType = ColumnType.RAW
            )
        }
        fun Int (name: String):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = Int.SIZE_BYTES,
                columnType = ColumnType.INT
            )
        }
        fun Long (name: String):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = Long.SIZE_BYTES,
                columnType = ColumnType.LONG
            )
        }

        fun Double (name: String):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = Double.SIZE_BYTES,
                columnType = ColumnType.DOUBLE
            )
        }
        fun String (name: String, length:Int=64):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = length,
                columnType = ColumnType.STRING
            )
        }
    }
    enum class ColumnType{
        INT,
        LONG,
        DOUBLE,
        STRING,
        RAW
    }
}