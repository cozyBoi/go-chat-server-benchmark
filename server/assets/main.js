window.addEventListener("load", function(evt) {
    var loc = window.location;
    var uri = 'http:';

    if (loc.protocol === 'https:') {
       uri = 'https:';
    }
    uri += '//' + loc.host;
    uri += loc.pathname;
    console.log(uri)

    //1. get room number
    httpGetAsync(loc.pathname + 'rooms', printRooms)
    httpGetAsync(loc.pathname + 'cookie', function(msg){
        console.log(msg)
        console.log(document.cookie)
    })

    document.getElementById("create").onclick = function(evt) {
        //=> post chat room
        var roomUri = uri + "rooms";
        console.log(roomUri);
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.open("POST", roomUri, true); // true for asynchronous 
        xmlHttp.send(null);
    };

    function enterRoom(evt) {
        var curr_id = evt.target.getAttribute('id');
        var roomUri = uri + "rooms/" + curr_id;
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function() { 
            document.location.href = roomUri;
        }
        xmlHttp.open("GET", roomUri, true); // true for asynchronous 
        xmlHttp.send(null);
    };

    function printRooms(room){
        //TODO add div
        console.log(room);
        var parseRooms = room.split(",");
        var i;
        for(i = 0; i < parseRooms.length; i++){
            let btn = document.createElement("button");
            btn.innerHTML = parseRooms[i];
            btn.onclick = enterRoom;
            btn.id = i+1;
            document.body.appendChild(btn);
        }
        //var newDiv = document.createElement("div");
        //newDiv.appendChild(btn);
    }

    function httpGetAsync(theUrl, callback)
    {
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function() { 
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
            callback(xmlHttp.responseText);
        }
        xmlHttp.open("GET", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }
});