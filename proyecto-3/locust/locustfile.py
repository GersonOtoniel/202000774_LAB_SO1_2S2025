from locust import HttpUser, task, between
import random

municipios = ["mixco","guatemala","amatitlan","chinautla"]
climas = ["sunny","cloudy","rainy","foggy"]

class WeatherUser(HttpUser):
    wait_time = between(1, 3)

    @task
    def send_tweet(self):
        data = {
            "municipality": random.randint(1,4),
            "temperature": random.randint(15, 35),
            "humidity": random.randint(30, 90),
            "weather": random.randint(1,4)
        }
        headers = {'Content-Type': 'application/json'}
        self.client.post("/weathertweet", json=data, headers=headers)