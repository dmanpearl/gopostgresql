function removeFromDb(item){
	console.log(`removeFromDb - item: '${item}'`)
	fetch(`/delete?item=${item}`, {method: "DELETE"}).then(res =>{
		if (res.status == 200){
			window.location.pathname = "/"
		}
	})
 }

function updateDb(item) {
	let input = document.getElementById(item)
	let newitem = input.value
	console.log(`updateDb - newItem: '${newitem}', item: '${item}'`)
	fetch(`/update?olditem=${item}&newitem=${newitem}`, {method: "PUT"}).then(res =>{
		if (res.status == 200){
			alert("Database updated")
			window.location.pathname = "/"
		}
	})
}
