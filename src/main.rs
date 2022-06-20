use std::fs;

fn main() {

    let filename: String = String::from("input.txt");

    let content = fs::read_to_string(filename)
        .expect("Something went wrong reading the file");

    let links: Vec<String> = content.split('\n')
                                    .map(|s| s.trim().to_string())
                                    .filter(|s| !s.is_empty())
                                    .collect::<Vec<_>>();

    for (index, link) in links.iter().enumerate() {
        println!("{}. {}", index+1, link);
    }
}
