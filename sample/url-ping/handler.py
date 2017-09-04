import requests

def print_url(url):
    try:
        r =  requests.get(url,timeout = 1)
        print(url +" => " + str(r.status_code))
    except:
        print("Timed out trying to reach URL.")

def handle(req):
    url = req.rstrip()
    print("Handle this -> " + url)
    if url.find("http") == -1:
        print("Give me a URL and I'll ping it for you.")
        return
    
    print_url(url)

# handle("http://faaster.io")
