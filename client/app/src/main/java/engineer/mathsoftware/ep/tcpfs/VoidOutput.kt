// Copyright (c) 2022 Tobias Briones. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// This file is part of https://github.com/tobiasbriones/ep-tcp-file-system

package engineer.mathsoftware.ep.tcpfs

import android.net.Uri

class VoidOutput() : Output {
    override fun updateUploadProgress(progress: Float) {
        TODO("Not yet implemented")
    }

    override fun uploadDone(file: String, chunksTotal: Int) {
        TODO("Not yet implemented")
    }

    override fun updateDownloadProgress(progress: Float) {
        TODO("Not yet implemented")
    }

    override fun downloadDone(
        data: ByteArray,
        uri: Uri,
        file: String,
        chunksTotal: Int
    ) {
        TODO("Not yet implemented")
    }
}
