wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["X-API-Key"] = "your_secret_api_key_here"  -- replace as needed
wrk.headers["X-Stream-ID"] = "your-test-stream-id"  -- replace with a valid stream ID

local payload = '{"key": "integration-test-value"}'
local path = "/stream/your-test-stream-id/send"  -- replace with the dynamic or test stream ID path

function request()
    return wrk.format(nil, path, nil, payload)
end
