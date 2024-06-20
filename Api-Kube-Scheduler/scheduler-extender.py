from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/filter', methods=['POST'])
def filter_nodes():
    data = request.get_json()
    nodes = data['nodes']['items']
    filtered_nodes = [node for node in nodes if node_meets_criteria(node)]
    return jsonify({'nodes': {'items': filtered_nodes}})

@app.route('/prioritize', methods=['POST'])
def prioritize_nodes():
    data = request.get_json()
    nodes = data['nodes']['items']
    scores = [{'name': node['metadata']['name'], 'score': calculate_score(node)} for node in nodes]
    return jsonify({'scores': scores})

def node_meets_criteria(node):
    return node['metadata']['labels'].get('kubernetes.io/arch') == 'arm64'

def calculate_score(node):
    for address in node['status']['addresses']:
        if address['type'] == 'InternalIP':
            ip = address['address']
            if int(ip.split('.')[-1]) % 2 == 0:
                return 10
    return 0

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=12345)