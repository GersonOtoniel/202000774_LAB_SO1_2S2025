use actix_web::{post, web, App, HttpResponse, HttpServer};
use serde::Deserialize;
use std::env;
use tonic::transport::Channel;

pub mod pb {
    tonic::include_proto!("weathertweet");
}

#[derive(Deserialize)]
struct TweetIn {
    municipality: String,
    temperature: i32,
    humidity: i32,
    weather: String,
}

fn map_muni(s: &str) -> i32 {
    match s {
        "mixco" => 1, "guatemala" => 2, "amatitlan" => 3, "chinautla" => 4, _ => 0
    }
}
fn map_weather(s: &str) -> i32 {
    match s {
        "sunny" => 1, "cloudy" => 2, "rainy" => 3, "foggy" => 4, _ => 0
    }
}

async fn grpc_client(addr: &str) -> pb::weather_tweet_service_client::WeatherTweetServiceClient<Channel> {
    pb::weather_tweet_service_client::WeatherTweetServiceClient::connect(format!("http://{}", addr))
        .await
        .expect("gRPC connect")
}

#[post("/api/tweet")]
async fn handle_tweet(payload: web::Json<TweetIn>) -> actix_web::Result<HttpResponse> {
    let use_kafka = true; // puedes decidir según header o query param
    let addr = if use_kafka {
        env::var("GRPC_KAFKA_ADDR").unwrap()
    } else {
        env::var("GRPC_RABBIT_ADDR").unwrap()
    };

    let mut client = grpc_client(&addr).await;
    let req = pb::WeatherTweetRequest {
        municipality: map_muni(&payload.municipality),
        temperature: payload.temperature,
        humidity: payload.humidity,
        weather: map_weather(&payload.weather),
    };

    let _resp = client.send_tweet(tonic::Request::new(req)).await
        .map_err(|e| actix_web::error::ErrorBadGateway(e.to_string()))?;

    Ok(HttpResponse::Ok().finish())
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenvy::dotenv().ok();
    let port: u16 = env::var("API_PORT").unwrap_or("8080".into()).parse().unwrap();
    HttpServer::new(|| App::new().service(handle_tweet))
        .bind(("0.0.0.0", port))?
        .run()
        .await
}