-- ===============================
-- WRK Script for Data Sending Benchmarking
-- ===============================

-- HTTP method and headers for data sending to a stream
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["X-API-Key"] = "your_secret_api_key_here"  -- Replace with actual API key
wrk.headers["X-Stream-ID"] = "your-test-stream-id"      -- Replace with an actual stream ID for testing

-- Payload and endpoint path
local payload = '{"key": "integration-test-value"}'
local path = "/stream/your-test-stream-id/send"         -- Update with actual stream ID in production

-- Function to format request for data sending
function request()
    -- Returns formatted POST request with dynamic path and payload
    return wrk.format(nil, path, nil, payload)
end
