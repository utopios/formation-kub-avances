from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/validate-pods', methods=['POST'])
def validate_pod():
    admission_review = request.get_json()
    pod = admission_review['request']['object']

    
    if 'owner' not in pod['metadata'].get('annotations', {}):
        return jsonify({
            "apiVersion": "admission.k8s.io/v1",
            "kind": "AdmissionReview",
            "response": {
                "uid": admission_review['request']['uid'],
                "allowed": False,
                "status": {
                    "message": "Missing 'owner' annotation."
                }
            }
        })
    
    return jsonify({
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "response": {
            "uid": admission_review['request']['uid'],
            "allowed": True
        }
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=443, ssl_context=('path/to/tls.crt', 'path/to/tls.key'))