from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/mutate', methods=['POST'])
def mutate():
    request_json = request.get_json()

    if 'request' in request_json and 'object' in request_json['request']:
        pod = request_json['request']['object']
        if pod['kind'] == 'Pod' and pod['metadata']['namespace'] == 'vote':
            annotations = pod['metadata'].get('annotations', {})
            annotations['mutated'] = 'true'
            patch = [
                {
                    "op": "add",
                    "path": "/metadata/annotations",
                    "value": annotations
                }
            ]

            import base64
            import json
            patch_base64 = base64.b64encode(json.dumps(patch).encode()).decode()

            return jsonify({
                'apiVersion': 'admission.k8s.io/v1',
                'kind': 'AdmissionReview',
                'response': {
                    'uid': request_json['request']['uid'],
                    'allowed': True,
                    'patchType': 'JSONPatch',
                    'patch': patch_base64
                }
            })

    return jsonify({
        'apiVersion': 'admission.k8s.io/v1',
        'kind': 'AdmissionReview',
        'response': {
            'uid': request_json['request']['uid'],
            'allowed': True
        }
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=443, ssl_context=('/etc/webhook/certs/tls.crt', '/etc/webhook/certs/tls.key'))