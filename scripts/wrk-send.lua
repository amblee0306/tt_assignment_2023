request = function()
	wrk.headers["Connection"] = "Keep-Alive"
	wrk.headers["Content-Type"] = "application/json"
	random_number = math.random(1,100000)
	wrk.body = '{"chat": "random1:a2", "text": "random number ' .. random_number .. '", "sender": "a2"}'
	path = "/api/send"
	return wrk.format("POST", path)
end
