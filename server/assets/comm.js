var ws;
window.addEventListener("load", function(evt) {
    //get last 30 chats
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    if (ws) {
        ws.close();
    }
    var loc = window.location;
    var uri = 'ws:';

    if (loc.protocol === 'https:') {
      uri = 'wss:';
    }
    uri += '//' + loc.host;
    uri += loc.pathname + '/ws';
    console.log(uri)
    
    ws = new WebSocket(uri);

    httpGetAsync(loc.pathname + '/chats', printChats)

    function printChats(chats){
        //TODO add div
        var i;
        
        if (chats == null || chats == "") {
            return
        }
        var parseChats = chats.split('\n');
        parseChats.forEach(x => {
            if(x == "") return
            let data = JSON.parse(x)
            if ("cid=" + data["sender"] == document.cookie){
                printR("Prev: " + data["msg"]);
            }
            else{
                printL("Prev: " + data["msg"]);
            }
        });
    }
    ws.onclose = function(evt) {
        ws = null;
    }
    ws.onmessage = function(evt) {
        printL("RESPONSE: " + evt.data);
    }
    ws.onerror = function(evt) {
        printL("ERROR: " + evt.data);
    }
    var printL = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        d.className = "bubbleleft"
        output.appendChild(d);  
        output.scroll(0, output.scrollHeight);
    };
    var printR = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        d.className = "bubbleright"
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        //input.value.className = "bubble"
        printR("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
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

window.addEventListener("beforeunload", function(evt) {
    ws = null;
});