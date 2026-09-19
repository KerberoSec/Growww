pub struct MmapWALSegment {
    pub segment_id: u64,
    pub capacity_bytes: usize,
    pub written_bytes: usize,
    pub buffer: Vec<u8>,
}

impl MmapWALSegment {
    pub fn new(segment_id: u64, capacity: usize) -> Self {
        Self {
            segment_id,
            capacity_bytes: capacity,
            written_bytes: 0,
            buffer: Vec::with_capacity(capacity),
        }
    }

    /// Append record to pre-allocated segment
    pub fn append_record(&mut self, record: &[u8]) -> Result<usize, String> {
        let rec_len = record.len();
        if self.written_bytes + rec_len + 4 > self.capacity_bytes {
            return Err("Segment capacity exceeded; rollover required".into());
        }

        // Write 4-byte length prefix
        let len_bytes = (rec_len as u32).to_le_bytes();
        self.buffer.extend_from_slice(&len_bytes);
        self.buffer.extend_from_slice(record);
        self.written_bytes += rec_len + 4;

        Ok(self.written_bytes)
    }
}
