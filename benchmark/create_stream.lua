-- ===============================
-- WRK Script for Stream Creation Benchmarking
-- ===============================

-- HTTP method and headers for stream creation
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["X-API-Key"] = "your_secret_api_key_here"  -- Replace with your actual API key

-- Function to format request for stream creation
function request()
    -- Return formatted POST request with necessary headers
    return wrk.format(nil)
end
