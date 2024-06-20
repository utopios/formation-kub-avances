const express = require('express');
const axios = require('axios');
const app = express();
const port = 3000;

const service1Url = 'http://service1.default.svc.cluster.local';
const service2Url = 'http://service2.default.svc.cluster.local';

app.get('/aggregate', async (req, res) => {
  try {
    const [userData, transactionData] = await Promise.all([
      axios.get(`${service1Url}/get`),
      axios.get(`${service2Url}`)
    ]);

    res.json({
      user: userData.data,
      transactions: transactionData.data
    });
  } catch (error) {
    res.status(500).send('Error during API aggregation');
  }
});

app.listen(port, () => {
  console.log(`API Aggregator running at http://localhost:${port}`);
});