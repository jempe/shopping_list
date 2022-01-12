var list = {
	heading : null,
	main : null,
	header : null,
	init : function() {
		this.heading = document.querySelector("h1");
		this.main = document.querySelector("main");
		this.header = document.querySelector("header");

		this.handle_route();

		window.addEventListener('hashchange', function()
		{
			window.list.handle_route();
		}, false);
	},
	get_all_lists : function() {
		fetch("/api/lists.json")
		.then(response => response.json())
		.then(data => this.render_lists(data))
	},
	render_lists: function(data) {
		this.heading.innerText = "Shopping Lists";
		this.header.innerHTML = "";
		this.main.innerHTML = "";

		let ul = document.createElement("ul");

		this.main.appendChild(ul);

		data.forEach(function(shopping_list)
		{
			let li = document.createElement("li");
			
			link = document.createElement("a");

			link.innerText = shopping_list.Name;
			link.href = "#lists/" + shopping_list.ID;

			li.appendChild(link);

			ul.appendChild(li);
		});
	},
	get_list : function(list_id) {
		fetch("/api/lists/" + list_id + ".json")
		.then(response => response.json())
		.then(data => this.render_list(data))
	},
	render_list: function(data) {
		this.heading.innerText = data.Name;
		this.header.innerHTML = "<a href=\"#\">Back</a>";
		this.main.innerHTML = "";

		let ul = document.createElement("ul");

		this.main.appendChild(ul);

		data.Items.forEach(function(item)
		{
			let li = document.createElement("li");
			
			let label = document.createElement("label");

			label.innerText = item.Name;
			
			let checkbox = document.createElement("input");
			checkbox.type = "checkbox";
			checkbox.checked = item.Checked;
			checkbox.dataset.id = item.ID;
			checkbox.setAttribute("onclick", "window.list.toggle_checkbox(this)");

			label.appendChild(checkbox);
			li.appendChild(label);
			ul.appendChild(li);
		});
	},
	toggle_checkbox : function(checkbox) {
		let checkbox_action = "uncheck";

		if(checkbox.checked)
		{
			checkbox_action = "check";
		}

		let action_url = "/api/items/" + checkbox.dataset.id + "/" + checkbox_action;

		fetch(action_url, {method : "PUT"});

	},
	handle_route : function() {
		let route = location.hash.replace(/^#/, "");

		if(/^lists\/[0-9]*$/.test(route))
		{
			let list_id = route.replace("lists/", "");
			this.get_list(list_id);
		}
		else
		{
			this.get_all_lists();
		}
	}
}

list.init();
