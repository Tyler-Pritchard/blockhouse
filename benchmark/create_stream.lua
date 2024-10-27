wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["X-API-Key"] = "your_secret_api_key_here" -- replace as needed

function request()
    return wrk.format(nil)
end
