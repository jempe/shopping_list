var socket = null;

if(window["WebSocket"])
{
	setup_websocket();
}
else
{
	console.error("Browser does not support websockets");
}

function setup_websocket()
{
	if(socket == null)
	{
		socket = new WebSocket("ws://" + location.host + "/remote");
		socket.onopen = function() {
			console.error("Websocket connection opened");
			document.querySelector("a.ws_connect").dataset.status = "connected";
			socket.send("update");
		}
		socket.onclose = function() {
			console.error("Websocket connection has been closed");
			socket = null;

			document.querySelector("a.ws_connect").dataset.status = "disconnected";
		}

		socket.onmessage = function(e) {
			if(e.data == "update")
			{
				window.list.handle_route();
			}
		}
	}
}
