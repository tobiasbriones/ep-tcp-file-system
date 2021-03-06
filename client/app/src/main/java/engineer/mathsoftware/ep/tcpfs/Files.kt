// Copyright (c) 2022 Tobias Briones. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// This file is part of https://github.com/tobiasbriones/ep-tcp-file-system

package engineer.mathsoftware.ep.tcpfs

import android.content.ContentResolver
import android.net.Uri
import java.io.ByteArrayOutputStream
import java.io.InputStream

const val SERVER_BUF_SIZE = 1024

fun read(res: ContentResolver, uri: Uri): ByteArray {
    val stream: InputStream? = res.openInputStream(uri)
    if (stream != null) {
        return getBytes(stream)
    }
    return ByteArray(0)
}

fun write(res: ContentResolver, uri: Uri, array: ByteArray) {
    val os = res.openOutputStream(uri)
    if (os != null) {
        os.write(array)
        os.close()
    }
}

private fun getBytes(inputStream: InputStream): ByteArray {
    val byteBuffer = ByteArrayOutputStream()
    val buffer = ByteArray(SERVER_BUF_SIZE)
    var len = 0
    while (inputStream.read(buffer)
            .also { len = it } != -1) {
        byteBuffer.write(buffer, 0, len)
    }
    return byteBuffer.toByteArray()
}
