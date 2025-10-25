use actix_web::{post, get, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
//use dotenvy::dotenv;
//use tonic::transport::Channel;

pub mod weathertweet {
    tonic::include_proto!("weathertweet");
}

use weathertweet::{
    weather_tweet_service_client::WeatherTweetServiceClient,
    WeatherTweetRequest, WeatherTweetResponse,
};

#[derive(Debug, Serialize, Deserialize)]
struct WeatherTweetInput {
    municipality: i32,
    temperature: i32,
    humidity: i32,
    weather: i32,
}

#[get("/")]
async fn health_check() -> impl Responder {
    HttpResponse::Ok().body("API is running")
}

#[post("/weathertweet")]
async fn create_weathertweet(
    data: web::Json<WeatherTweetInput>,
) -> impl Responder {
    println!("Received request: {:?}", data);
    let _ = dotenvy::dotenv().ok();
    let grpc_url= std::env::var("GRPC_SERVER_URL").unwrap_or_else(|_| "http://0.0.0.0:50051".to_string());
    let mut client = match WeatherTweetServiceClient::connect(grpc_url)
        .await {
            Ok(client) => client,
            Err(e) => {
                return HttpResponse::InternalServerError().body(format!("Error connecting to gRPC server: {}", e));
            }
        };
    let request = tonic::Request::new(WeatherTweetRequest {
        municipality: data.municipality,
        temperature: data.temperature,
        humidity: data.humidity,
        weather: data.weather,
    });
    
    match client.send_weather_tweet(request).await {
        Ok(response) => {
            let response_inner: WeatherTweetResponse = response.into_inner();
            HttpResponse::Ok().json(response_inner)
        }
        Err(e) => {
            println!("gRPC request failed: {}", e);
            HttpResponse::InternalServerError().body(format!("gRPC request failed: {}", e))
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenvy::dotenv().ok();
    //env_logger::init();
    println!("Starting API server on http://0.0.0.0:8080");
    HttpServer::new(|| {
        App::new()
            .service(health_check)
            .service(create_weathertweet)
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
