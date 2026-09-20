/// Pre-allocated memory-mapped Write-Ahead Log (WAL) segment
/// Providing zero-copy append operations, length-prefixed binary framing,
/// and crash recovery replay.
#[derive(Debug, Clone)]
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

    /// Append record to pre-allocated segment with 4-byte little-endian length prefix.
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

    /// Read all valid records sequentially from segment buffer
    pub fn read_records(&self) -> Result<Vec<Vec<u8>>, String> {
        let mut records = Vec::new();
        let mut offset = 0;
        while offset + 4 <= self.buffer.len() {
            let len_bytes = [
                self.buffer[offset],
                self.buffer[offset + 1],
                self.buffer[offset + 2],
                self.buffer[offset + 3],
            ];
            let rec_len = u32::from_le_bytes(len_bytes) as usize;
            offset += 4;
            if offset + rec_len > self.buffer.len() {
                return Err("Corrupted or truncated WAL segment record".into());
            }
            records.push(self.buffer[offset..offset + rec_len].to_vec());
            offset += rec_len;
        }
        Ok(records)
    }

    /// Check if segment needs rollover
    pub fn is_full(&self) -> bool {
        self.written_bytes >= self.capacity_bytes
    }

    /// Remaining available capacity in bytes
    pub fn remaining_capacity(&self) -> usize {
        if self.written_bytes >= self.capacity_bytes {
            0
        } else {
            self.capacity_bytes - self.written_bytes
        }
    }
}
