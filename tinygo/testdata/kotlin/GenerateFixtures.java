// Generates the Kotlin fixtures consumed by the Go compatibility tests.
// See ../README.md for the compile and run commands.
import com.qapsoft.io.BinaryColumn;
import com.qapsoft.io.BinaryList;
import com.qapsoft.io.BinaryStream;
import com.qapsoft.io.BinaryTableData;
import com.qapsoft.io.BinaryTableHeader;
import com.qapsoft.io.ByteArrayStreamAutoSize;
import com.qapsoft.io.ByteArrayStreamReader;
import com.qapsoft.io.DeflateBinaryStreamWriterKt;
import com.qapsoft.io.IoHelperKt;
import com.qapsoft.util.AESEncryptor;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;

public final class GenFixtures {
    public static void main(String[] args) throws Exception {
        Path dir = Paths.get(args[0]);
        Files.createDirectories(dir);

        BinaryColumn[] cols = new BinaryColumn[]{
                BinaryColumn.Companion.Int("_id"),
                BinaryColumn.Companion.String("first_name", 16),
                BinaryColumn.Companion.String("last_name", 16),
                BinaryColumn.Companion.Long("created"),
                BinaryColumn.Companion.Double("score"),
                BinaryColumn.Companion.Raw("payload", 8)
        };
        BinaryTableHeader header = new BinaryTableHeader(16, Arrays.asList(cols));

        // Growable in-memory table.
        ByteArrayStreamAutoSize memory = new ByteArrayStreamAutoSize();
        BinaryTableData table = new BinaryTableData(header, (BinaryStream) memory);
        table.setValue(0, "_id", IoHelperKt.toByteArray(7));
        table.setValue(0, "first_name", "Alice".getBytes(StandardCharsets.UTF_8));
        table.setValue(0, "last_name", "Saleh".getBytes(StandardCharsets.UTF_8));
        table.setValue(0, "created", IoHelperKt.toByteArray(1700000000000L));
        table.setValue(0, "score", IoHelperKt.toByteArray(12.5d));
        table.setValue(0, "payload", new byte[]{1, 2, 3});
        table.setValue(3, "_id", IoHelperKt.toByteArray(99));
        table.setValue(3, "first_name", "Maged".getBytes(StandardCharsets.UTF_8));
        Files.write(dir.resolve("table.bin"), memory.toByteArray());

        // Fixed-capacity Kotlin file table.
        File file = dir.resolve("table-file.bin").toFile();
        if (file.exists()) {
            file.delete();
        }
        BinaryTableData fileTable = new BinaryTableData(header, file);
        fileTable.setValue(3, "_id", IoHelperKt.toByteArray(99));
        fileTable.setValue(3, "first_name", "Maged".getBytes(StandardCharsets.UTF_8));
        ((java.io.Closeable) fileTable.getStreamWriter()).close();

        // Binary list.
        ByteArrayStreamAutoSize listStream = new ByteArrayStreamAutoSize();
        BinaryList list = new BinaryList(listStream, 8);
        list.add("maged".getBytes(StandardCharsets.UTF_8));
        list.add("ahmed".getBytes(StandardCharsets.UTF_8));
        list.add(new byte[]{0, 1, 0, (byte) 255, 0});
        Files.write(dir.resolve("list.bin"), listStream.toByteArray());

        // Deflate stream and its original data.
        byte[] raw = new byte[20000];
        for (int i = 0; i < raw.length; i++) {
            raw[i] = (byte) (i % 64);
        }
        ByteArrayStreamAutoSize deflated = new ByteArrayStreamAutoSize();
        DeflateBinaryStreamWriterKt.deflate(new ByteArrayStreamReader(raw), deflated, 1024, 0);
        Files.write(dir.resolve("deflate.bin"), deflated.toByteArray());
        Files.write(dir.resolve("deflate-raw.bin"), raw);

        // AES/CBC/PKCS5 fixture.
        byte[] plain = "BinaryTableData AES compatibility fixture".getBytes(StandardCharsets.UTF_8);
        byte[] iv = new byte[16];
        for (int i = 0; i < iv.length; i++) {
            iv[i] = (byte) i;
        }
        ByteArrayOutputStream encrypted = new ByteArrayOutputStream();
        AESEncryptor.INSTANCE.encrypt("secret-key", iv, new ByteArrayInputStream(plain), encrypted);
        Files.write(dir.resolve("aes.bin"), encrypted.toByteArray());
        Files.write(dir.resolve("aes-plain.bin"), plain);
        Files.write(dir.resolve("aes-iv.bin"), iv);
    }
}
