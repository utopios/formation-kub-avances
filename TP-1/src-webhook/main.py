from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route("/convert", methods=["POST"])
def convert():
    req = request.get_json()
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

    desired_api_version = req['request']['desiredAPIVersion']

    for obj in req['request']['objects']:
        if desired_api_version == "myorg.com/v1":
            # Assume obj is in v2 format and needs to be converted to v1
            converted_obj = obj.copy()  # Make a shallow copy to preserve original data
            # Conversion logic from v2 to v1
            converted_obj['apiVersion'] = "myorg.com/v1"
            converted_obj['spec']['config'] = converted_obj['spec'].pop('configuration')
            response['response']['convertedObjects'].append(converted_obj)
        elif desired_api_version == "myorg.com/v2":
            # Assume obj is in v1 format and needs to be converted to v2
            converted_obj = obj.copy()  # Make a shallow copy to preserve original data
            # Conversion logic from v1 to v2
            converted_obj['apiVersion'] = "myorg.com/v2"
            converted_obj['spec']['configuration'] = converted_obj['spec'].pop('config')
            response['response']['convertedObjects'].append(converted_obj)

    return jsonify(response)

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=443, ssl_context=('/etc/ssl/certs/webhook_server.crt', '/etc/ssl/private/webhook_server.key'))
