/**
 * client.js
 * 
 * An atmosphere node client program to interactively calls the greeter service
 * 
 * Usages: 
 * node client.js [options] [greeter-url]
 *
 * Options:
 *   -u    user
 *   -p    password
 *   -k    use insecure mode in TLS
 *   -v    enable verbose mode
 *
 * Usage Samples:
 *  node client.js http://localhost:8091/samples/greeter
 *  node client.js -t true http://localhost:8091/samples/greeter
 *  node client.js -k -u 'guest' https://localhost:8495/samples/greeter
 * 
 */

"use strict";

const HOST_URL = 'http://localhost:8091/samples/greeter';

var hosturl = HOST_URL;
var authuser
var authpassword
var insecure;
var trace = false;

var arg;
for (var i = 2; i < process.argv.length; i++) {
    arg = process.argv[i];
    if (arg === "-u") {
        authuser = process.argv[++i];
        arg = undefined;
    } else if (arg === "-p") {
        authpassword = process.argv[++i];
        arg = undefined;
    } else if (arg === "-k") {
        insecure = true;
        arg = undefined;
    } else if (arg === "-v") {
        trace = true;
        arg = undefined;
    }
}
if (arg != undefined) {
    hosturl = arg;
}

console.log("Host URL: " + hosturl + (authuser != undefined ? " (Basic" : " (No") + " Authentication)")

var reader = require('readline');
var writable = require('stream').Writable;
var mutableStdout = new writable({
    write: function(chunk, encoding, callback) {
        if (!this.muted)
            process.stdout.write(chunk, encoding);
        callback();
    }
});
var prompt = reader.createInterface({input: process.stdin, output: mutableStdout, terminal: true});
mutableStdout.muted = false;

var headers = {'X-Requested-With': 'XMLHttpRequest'}
if (authuser !== undefined) {
    if (authpassword === undefined) {
        prompt.question("Enter host password for user '" + authuser + "':", function(pw) {
            authpassword = pw;
            mutableStdout.muted = false;
            console.log();
            prompt.prompt();
            headers["Authorization"] = "Basic " + Buffer.from(authuser+":"+authpassword).toString("base64")
            queryUser();
        });
        mutableStdout.muted = true;
    } else {
        headers["Authorization"] = "Basic " + Buffer.from(authuser+":"+authpassword).toString("base64")
        queryUser();
    }
} else {
    queryUser();
}

const swagsock = require('swagsock.js');
var isopen = false;

// greeter's consle commands
const COMMAND_LIST = 
    [["ping",       "Ping", "ping"],
     ["echo",       "Echo the message", "echo text"],
     ["greet",      "Greet the person", "greet name text"],
     ["status",     "Display the greeting status", "status name"],
     ["summary",    "Display the greeting summary", "summary"],
     ["subscribe",  "Subscribe to the greeting events", "subscribe"],
     ["unsubscribe","Unsubscribe from the greeting events", "unsubscribe sid"],
     ["help",       "Display this help message", "help"],
     ["quit",       "Quit the application", "quit"]];


/////////////// utilities

function selectOption(c, opts) {
    var i = c.length == 0 ? 0 : parseInt(c);
    if (!(i >= 0 && i < opts.length)) {
        console.log('Invalid selection: ' + c + '; Using ' + opts[0]);
        i = 0;
    }
    return opts[i];
}

function getArgs(name, msg, num) {
    var sp = name.length;
    if (msg.length > name.length && msg.charAt(name.length) != ' ') {
        // remove the command suffix
        sp = msg.indexOf(' ', name.length);
        if (sp < 0) {
            sp = msg.length;
        }
    }
    var v = msg.substring(sp).trim();
    if (v.length > 0) {
        if (num == 0) {
            // return as a single value
            return v;
        } else {
            // return as an array of num + 1 elements containing the first num arguments and the rest
            var params = [];
            while (num > 0) {
                var param;
                if (num == 1) {
                    params.push(v);
                } else {
                    var p = v.indexOf(' ', 0);
                    if (p < 0) {
                        break;
                    }
                    param = v.substring(0, p);
                    v = v.substring(p + 1).trim();
                }
                params.push(param);
                num--;
            }
            if (num > 1) {
                throw "Invalid arguments for " + name;
            }
            return params;
        }
    }
}

function errorUsage(v, msg) {
    console.log("Error: Missing arguments for " + v);
    for (var i = 0; i < COMMAND_LIST.length; i++) { 
        var c = COMMAND_LIST[i][0];
        if (v == c) {
            console.log("Usage: " + COMMAND_LIST[i][2]);
        }
    }
}

function queryUser() {
    console.log("Choose user ...");
    prompt.setPrompt("user: ", 6);
    prompt.prompt();
}

/////////////// commands

function doHelp(v) {
    if (!v) {
        console.log('Available commands');
        for (var i = 0; i < COMMAND_LIST.length; i++) { 
            var c = COMMAND_LIST[i][0];
            console.log(c + "                    ".substring(0, 20 - c.length) + COMMAND_LIST[i][1]);
        }
    } else {
        var found = false;
        for (var i = 0; i < COMMAND_LIST.length; i++) { 
            var c = COMMAND_LIST[i][0];
            if (v == c) {
                console.log(COMMAND_LIST[i][1]);
                console.log("Usage: " + COMMAND_LIST[i][2]);
                found = true;
            }
        }
        if (!found) {
            throw "Uknown command: " + v;
        }
    }
}

function toConsole(header, content) {
    console.log("res"+header.id+":", JSON.stringify(header) + content);
    prompt.setPrompt(userprompt, 2);
    prompt.prompt();
}

function doGreet(v) {
    if (!v) {
        errorUsage("greet");
        return;
    }
    var req = swagSocket.request().pathpattern("/v1/greet/{name}").pathparam("name", user).method("POST").content(JSON.stringify({ "name": v[0], "text": v[1]}), "application/json");
    swagSocket.send(req, toConsole);
}

function doStatus(v) {
    if (!v) {
        errorUsage("status");
        return;
    }
    var req = swagSocket.request().pathpattern("/v1/greet/{name}").pathparam("name", v[0]);
    swagSocket.send(req, toConsole);
}

function doSummary(v) {
    var req = swagSocket.request().pathpattern("/v1/greet");
    swagSocket.send(req, toConsole);
}

function doSubscribe(v) {
    var req = swagSocket.request().pathpattern("/v1/subscribe/{name}").pathparam("name", user).subscribe(true);
    swagSocket.send(req, toConsole);
}

function doUnsubscribe(v) {
    if (!v) {
        errorUsage("unsubscribe");
        return;
    }

    var req = swagSocket.request().pathpattern("/v1/unsubscribe/{sid}").pathparam("sid", v[0]).method("DELETE").unsubscribe(v[0]);
    swagSocket.send(req, toConsole);
}

function doPing(v) {
    var req = swagSocket.request().pathpattern("/v1/ping");
    swagSocket.send(req, toConsole);
}

function doEcho(v) {
    if (!v) {
        errorUsage("echo");
        return;
    }
    var req = swagSocket.request().pathpattern("/v1/echo").method("POST").content(v[0], "text/plain");
    swagSocket.send(req, toConsole);
}


function doQuit() {
    swagSocket.close();
    process.exit(0);
}

////////////////////

var transport = null;
var swagSocket = null;

var user = null;
var userprompt = "> ";
prompt.
on('line', function(line) {
    try {
        var msg = line.trim();
        if (user == null) {
            user = msg
            userprompt = user + userprompt;
            swagSocket = swagsock.swaggersocket(hosturl);
            swagSocket.on("open", function() {
                console.log('Connected using websocket');
                doHelp();
                prompt.setPrompt(userprompt, 2);
                prompt.prompt();
            });
            swagSocket.on("close", function() {
                console.log('Connection closed. Reconnecting ...');
                prompt.setPrompt(userprompt, 2);
                prompt.prompt();
            });
            swagSocket.open();
            console.log("Connecting using websocket ...");
        } else if (msg.length == 0) {
//            doHelp();
        } else if (msg.indexOf("ping") == 0) {
            doPing();
        } else if (msg.indexOf("echo") == 0) {
            doEcho(getArgs("echo", msg, 1));
        } else if (msg.indexOf("greet") == 0) {
            doGreet(getArgs("greet", msg, 2));
        } else if (msg.indexOf("help") == 0) {
            doHelp(getArgs("help", msg, 0));
        } else if (msg.indexOf("status") == 0) {
            doStatus(getArgs("status", msg, 1));
        } else if (msg.indexOf("summary") == 0) {
            doSummary();
        } else if (msg.indexOf("subscribe") == 0) {
            doSubscribe();
        } else if (msg.indexOf("unsubscribe") == 0) {
            doUnsubscribe(getArgs("unsubscribe", msg, 1));
        } else if (msg.indexOf("quit") == 0) {
            doQuit();
        } else {
            throw "Uknown command: " + msg;
        }
    } catch (err) {
        console.log("Error: " + err);
    }
    prompt.setPrompt(userprompt, 2);
    prompt.prompt();
}).

on('close', function() {
    doQuit();
});
