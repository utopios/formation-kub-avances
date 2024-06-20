from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/validate', methods=['POST'])
def validate():
    request_info = request.get_json()
    uid = request_info['request']['uid']
    allowed = True
    message = "Pod validation successful"

    # Exemple de validation : vérifier si le pod contient un label 'app: vote'
    labels = request_info['request']['object']['metadata'].get('labels', {})
    if 'app' not in labels or labels['app'] != 'vote':
        allowed = False
        message = "Missing required label: 'app: vote'"

    # Créer une réponse AdmissionReview
    response = {
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "response": {
            "uid": uid,
            "allowed": allowed,
            "status": {
                "message": message
            }
        }
    }
    return jsonify(response)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=443, ssl_context=('/etc/webhook/certs/server.crt', '/etc/webhook/certs/server.key'))