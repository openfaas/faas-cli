import requests

def print_url(url):
    r =  requests.get(url)
    print(url +" => " + str(r.status_code))

def handle(req):
    print("Handle this -> " + req)
    if req.find("http") == -1:
        print("Give me a URL and I'll ping it for you.")
        return
    
    print_url(req)
