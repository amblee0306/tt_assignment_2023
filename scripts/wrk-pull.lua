request = function()
	wrk.headers["Connection"] = "Keep-Alive"
	wrk.headers["Content-Type"] = "application/json"
	-- random_number = math.random(1,3)
	wrk.body = '{"chat": "a1:a2", "cursor": 0, "limit": 10}'
	path = "/api/pull"
	return wrk.format("GET", path)
end
