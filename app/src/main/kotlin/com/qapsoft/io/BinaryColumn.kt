package com.qapsoft.io

data class BinaryColumn (
    val name:String,
    val length:Int
){
    companion object{
        fun Raw (name: String, length: Int):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = length
            )
        }
        fun Int (name: String):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = Int.SIZE_BYTES
            )
        }
        fun Long (name: String):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = Long.SIZE_BYTES
            )
        }

        fun Double (name: String):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = Double.SIZE_BYTES
            )
        }
        fun String (name: String, length:Int=64):BinaryColumn{
            return BinaryColumn(
                name = name,
                length = length
            )
        }
    }
}