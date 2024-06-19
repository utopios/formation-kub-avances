from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route("/convert", methods=["POST"])
def convert():
    req = request.get_json()

        ### Logique de conversion
    
    response = {
        "apiVersion": "apiextensions.k8s.io/v1",
        "kind": "ConversionReview",
        "response": {
            "uid": req['request']['uid'],
            "convertedObjects": [],
            "result": {
                "status": "Success",
                "message": "Conversion réussie"
            }
        }
    }
    return jsonify(response)

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=443, ssl_context=('/etc/ssl/certs/webhook_server.crt', '/etc/ssl/private/webhook_server.key'))