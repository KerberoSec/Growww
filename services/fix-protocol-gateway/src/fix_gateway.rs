use std::collections::HashMap;

/// FIX 5.0 SP2 message tag constants
pub const TAG_BEGIN_STRING: u32 = 8;
pub const TAG_MSG_TYPE: u32 = 35;
pub const TAG_SENDER_COMP_ID: u32 = 49;
pub const TAG_TARGET_COMP_ID: u32 = 56;
pub const TAG_MSG_SEQ_NUM: u32 = 34;
pub const TAG_CL_ORD_ID: u32 = 11;
pub const TAG_SYMBOL: u32 = 55;
pub const TAG_SIDE: u32 = 54;
pub const TAG_ORDER_QTY: u32 = 38;
pub const TAG_PRICE: u32 = 44;

#[derive(Debug, Clone)]
pub struct FIXMessage {
    pub msg_type: String,
    pub fields: HashMap<u32, String>,
}

pub struct FIX50SP2Gateway {
    pub comp_id: String,
    pub expected_inbound_seq: u64,
    pub outbound_seq: u64,
}

impl FIX50SP2Gateway {
    pub fn new(comp_id: &str) -> Self {
        Self {
            comp_id: comp_id.to_string(),
            expected_inbound_seq: 1,
            outbound_seq: 1,
        }
    }

    /// Parse standard SOH (\x01) delimited FIX tag-value stream
    pub fn parse_fix_message(&mut self, raw: &str) -> Result<FIXMessage, String> {
        let mut fields = HashMap::new();
        for kv in raw.split('\x01') {
            if kv.is_empty() {
                continue;
            }
            let mut parts = kv.splitn(2, '=');
            if let (Some(tag_str), Some(val)) = (parts.next(), parts.next()) {
                if let Ok(tag) = tag_str.parse::<u32>() {
                    fields.insert(tag, val.to_string());
                }
            }
        }

        let msg_type = match fields.get(&TAG_MSG_TYPE) {
            Some(t) => t.clone(),
            None => return Err("Missing Tag 35 MsgType".into()),
        };

        if let Some(seq_str) = fields.get(&TAG_MSG_SEQ_NUM) {
            if let Ok(seq) = seq_str.parse::<u64>() {
                if seq != self.expected_inbound_seq {
                    // In production, initiate ResendRequest (MsgType 2)
                }
                self.expected_inbound_seq = seq + 1;
            }
        }

        Ok(FIXMessage { msg_type, fields })
    }

    /// Build FIX Execution Report (MsgType 8)
    pub fn build_execution_report(
        &mut self,
        cl_ord_id: &str,
        exec_id: &str,
        symbol: &str,
        side: &str,
        qty: f64,
        price: f64,
    ) -> String {
        let seq = self.outbound_seq;
        self.outbound_seq += 1;

        format!(
            "8=FIX.5.0SP2\x0135=8\x0149={}\x0134={}\x0111={}\x0117={}\x0155={}\x0154={}\x0138={:.4}\x0144={:.2}\x01150=2\x0139=2\x0110=000\x01",
            self.comp_id, seq, cl_ord_id, exec_id, symbol, side, qty, price
        )
    }
}
