import json
import requests


def main():
    infile = open("/home/salazar/data.json", "r")
    line = infile.readline()
    full = json.loads(line)

    for entry in full:
        try: 
            roll = int(entry["i"])
        except:
            continue
        if roll < 180000 or roll > 202000:
            continue
        ok = dict()
        ok["roll"] = roll
        ok["name"] = entry["n"]
        ok["email"] = entry["u"]+"@iitk.ac.in"
        ok["password"] = entry["u"]
        dt = json.dumps(ok)
        r = requests.post("http://localhost:8080/signup", data=dt)
        if r.status_code != 200:
            print(r.content)    

if __name__ == '__main__':
    main()
