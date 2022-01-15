var socket = null;

if(window["WebSocket"])
{
	socket = new WebSocket("ws://" + location.host + "/remote");
	socket.onclose = function() {
		console.error("Websocket connection has been closed");
	}

	socket.onmessage = function(e) {
		if(e.data == "update")
		{
			window.list.handle_route();
		}
	}
}
else
{
	console.error("Browser does not support websockets");
}


