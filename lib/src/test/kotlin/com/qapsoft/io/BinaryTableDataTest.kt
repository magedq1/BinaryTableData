package com.qapsoft.io

import com.qapsoft.util.EncryptionHelper
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test
import kotlin.random.Random

class BinaryTableDataTest{
    @Test
    fun test(){
        //this required a lot of memory
        val rand = Random.Default
        val rows = rand.nextInt(100, 500)
        val col = rand.nextInt(1, 30)
        val schemaSize = rand.nextInt(1024, 64*1024)

        var primarySize = 0
        val columns:List<BinaryColumn> = mutableListOf<BinaryColumn>().also { cols->
            repeat(col){i->
                val colLength = rand.nextInt(1,500)
                primarySize+=colLength
                cols.add(
                    BinaryColumn(
                        EncryptionHelper.randomString(rand.nextInt(1,10)),
                        colLength
                    )
                )
            }
        }

        val header = BinaryTableHeader(
            maxRowsCount = rows,
            columns = columns
        )

        val writer = ByteArrayStreamWriterAutoSize()
        var table = BinaryTableData(
            header = header,
            streamWriter = writer
        )
        val map = mutableMapOf<String,ByteArray>()
        repeat(rows){rowIndex->
            columns.forEachIndexed { colIndex, c ->
                val data = EncryptionHelper.randomString(rand.nextInt(c.length)).toByteArray()
                table.setValue(rowIndex, colIndex, data)

                if(map.size<20 && rand.nextInt(20)<5){
                    map["$rowIndex-$colIndex"]=data
                }
            }
        }

        table = BinaryTableData.from(
            ByteArrayStreamReader(
                writer.toByteArray()
            )
        )

        map.forEach { entry->
            val sepPos = entry.key.indexOf('-')
            val rowId = entry.key.substring(0,sepPos).toInt()
            val colId = entry.key.substring(sepPos+1).toInt()
            assertArrayEquals(
                entry.value,
                table.getValue(rowId, colId).copyOfRange(0, entry.value.size)
                )
        }
    }
}