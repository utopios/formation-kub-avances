from flask import Flask, request, jsonify
import json
import base64

app = Flask(__name__)

@app.route('/mutate-pods', methods=['POST'])
def mutate_pod():

    admission_review = request.get_json()
    pod = admission_review['request']['object']
    namespace = admission_review['request']['namespace']

   
    department_label = {'finance': 'Finance', 'tech': 'Technology'}.get(namespace, 'General')

    patch = [{
        "op": "add",
        "path": "/metadata/labels/department",
        "value": department_label
    }]
    patch_response = base64.b64encode(json.dumps(patch).encode()).decode()

    return jsonify({
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "response": {
            "uid": admission_review['request']['uid'],
            "allowed": True,
            "patchType": "JSONPatch",
            "patch": patch_response
        }
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=443, ssl_context=('path/to/tls.crt', 'path/to/tls.key'))
