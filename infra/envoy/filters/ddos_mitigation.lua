-- ==============================================================================
-- Growww Sovereign Exchange - Envoy Perimeter DDoS & Malicious Flow Filter
-- ==============================================================================

local BLOCKED_USER_AGENTS = {
    "sqlmap",
    "nikto",
    "masscan",
    "gobuster",
    "dirbuster",
    "havij",
    "acunetix",
    "wpscan"
}

function envoy_on_request(handle)
    local headers = handle:headers()
    local path = headers:get(":path") or ""
    local user_agent = string.lower(headers:get("user-agent") or "")
    local host = headers:get(":authority") or headers:get("host") or ""

    -- Check 1: Enforce presence of Host / Authority header (RFC 7230 / HTTP/2)
    if host == "" then
        handle:respond(
            {[":status"] = "400", ["content-type"] = "application/json"},
            '{"error": "Bad_Request", "message": "Missing Host or Authority header", "code": 400}\n'
        )
        return
    end

    -- Check 2: Block known automated attack scanners & exploit bots
    for _, blocked in ipairs(BLOCKED_USER_AGENTS) do
        if string.find(user_agent, blocked, 1, true) then
            handle:logWarn(string.format("DDoS Mitigation: Blocked automated scanner User-Agent '%s'", user_agent))
            handle:respond(
                {[":status"] = "403", ["content-type"] = "application/json", ["x-firewall-action"] = "blocked"},
                '{"error": "Access_Forbidden", "message": "Automated security scanner detected and dropped", "code": 403}\n'
            )
            return
        end
    end

    -- Check 3: Clock Skew & Replay Attack Defense on Trading Routes
    if string.find(path, "/api/v1/orders", 1, true) then
        local timestamp_str = headers:get("x-request-timestamp")
        if timestamp_str then
            local client_time = tonumber(timestamp_str)
            local current_time = os.time()
            if client_time and math.abs(current_time - client_time) > 300 then
                handle:logWarn(string.format("DDoS Mitigation: Replay attack detected. Skew: %d seconds", math.abs(current_time - client_time)))
                handle:respond(
                    {[":status"] = "401", ["content-type"] = "application/json"},
                    '{"error": "Unauthorized_Timestamp_Skew", "message": "Request timestamp skew exceeds 300 seconds", "code": 401}\n'
                )
                return
            end
        end
    end

    -- Check 4: Rate-burst penalty header injection for suspicious request signatures
    if string.len(path) > 1024 then
        handle:respond(
            {[":status"] = "414", ["content-type"] = "application/json"},
            '{"error": "URI_Too_Long", "message": "Request URI exceeds maximum permissible length", "code": 414}\n'
        )
        return
    end
end
