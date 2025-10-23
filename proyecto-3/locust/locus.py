from locust import HttpUser, task, between
import random

municipios = ["mixco","guatemala","amatitlan","chinautla"]
climas = ["sunny","cloudy","rainy","foggy"]

class WeatherUser(HttpUser):
    wait_time = between(1, 3)

    @task
    def send_tweet(self):
        data = {
            "municipality": random.choice(municipios),
            "temperature": random.randint(15, 35),
            "humidity": random.randint(30, 90),
            "weather": random.choice(climas)
        }
        self.client.post("/api/tweet", json=data)