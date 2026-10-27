// Reads fixtures written by the Go port and verifies them with the original
// Kotlin implementation. See ../README.md for the compile and run commands.
import com.qapsoft.io.BinaryList;
import com.qapsoft.io.BinaryTableData;
import com.qapsoft.io.ByteArrayStream;
import com.qapsoft.io.ByteArrayStreamReader;
import com.qapsoft.io.DeflateBinaryStreamReader;
import com.qapsoft.io.IoHelperKt;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;

public final class VerifyGoFixtures {
    public static void main(String[] args) throws Exception {
        Path dir = Paths.get(args[0]);
        byte[] tableBytes = Files.readAllBytes(dir.resolve("go-table.bin"));
        BinaryTableData table = BinaryTableData.Companion.from(tableBytes);
        check(IoHelperKt.getInt(table, 0, "_id") == 7, "int");
        check(IoHelperKt.getString(table, 0, "first_name", java.nio.charset.StandardCharsets.UTF_8).equals("Alice"), "string");
        check(IoHelperKt.getString(table, 0, "last_name", java.nio.charset.StandardCharsets.UTF_8).equals("Saleh"), "string2");
        check(IoHelperKt.getLong(table, 0, "created") == 1700000000000L, "long");
        check(IoHelperKt.getDouble(table, 0, "score") == 12.5d, "double");
        byte[] payload = table.getValue(0, "payload");
        check(payload[0] == 1 && payload[1] == 2 && payload[2] == 3 && payload[3] == 0, "raw");
        check(IoHelperKt.getInt(table, 3, "_id") == 99, "row3 int");
        check(IoHelperKt.getString(table, 3, "first_name", java.nio.charset.StandardCharsets.UTF_8).equals("Maged"), "row3 string");

        BinaryTableData fileTable = BinaryTableData.Companion.from(dir.resolve("go-table-file.bin").toFile());
        check(IoHelperKt.getInt(fileTable, 3, "_id") == 99, "file int");
        check(IoHelperKt.getString(fileTable, 3, "first_name", java.nio.charset.StandardCharsets.UTF_8).equals("Maged"), "file string");
        fileTable.getStreamWriter();

        byte[] listBytes = Files.readAllBytes(dir.resolve("go-list.bin"));
        BinaryList list = new BinaryList(new ByteArrayStream(listBytes, false), 8);
        check(list.getSize() == 3, "list size");
        check(Arrays.equals(list.getBytes(0), "maged".getBytes()), "list 0");
        check(Arrays.equals(list.getBytes(1), "ahmed".getBytes()), "list 1");
        check(Arrays.equals(list.getBytes(2), new byte[]{0, 1, 0, (byte) 255, 0}), "list 2");

        byte[] deflated = Files.readAllBytes(dir.resolve("go-deflate.bin"));
        byte[] raw = Files.readAllBytes(dir.resolve("go-deflate-raw.bin"));
        DeflateBinaryStreamReader reader = new DeflateBinaryStreamReader(new ByteArrayStreamReader(deflated), 10);
        byte[] inflated = IoHelperKt.getBytesAt(reader, 0, (int) reader.length());
        check(Arrays.equals(inflated, raw), "deflate");

        System.out.println("Kotlin verified all Go fixtures");
    }

    private static void check(boolean condition, String label) {
        if (!condition) {
            throw new AssertionError("failed: " + label);
        }
    }
}
