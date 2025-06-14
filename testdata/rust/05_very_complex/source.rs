// 05_very_complex.rs
// A test for procedural macros, GATs, and advanced generic concepts.

// --- Assume this proc-macro is defined in a separate crate ---
// use my_proc_macros::Configurable;
// To make this file self-contained, we define a dummy trait.
// A real parser would see `#[derive(Configurable)]` and need to handle it.
pub trait Configurable {
    fn from_source<S: DataSource>(source: &S) -> Self;
}
// --- End of proc-macro simulation ---

/// A trait with a Generic Associated Type (GAT).
/// `Reader` has its own lifetime `'a`, which is tied to `&'a self`.
pub trait DataSource {
    type Reader<'a>: std::io::Read where Self: 'a;
    type Config<T: Clone>: Clone;
    
    fn get_reader<'a>(&'a self) -> Self::Reader<'a>;
    fn get_config<T: Clone>(&self) -> Self::Config<T>;
    
    /// Method with complex generic bounds
    fn process_with_bounds<'a, T, U>(&'a self, input: T) -> U 
    where
        T: AsRef<str> + 'a,
        U: From<&'a str> + Default,
        Self::Reader<'a>: std::io::BufRead;
}

/// A struct that would use a procedural derive macro in a real project.
/// The parser must handle attributes on the struct and its fields.
#[derive(Debug, Clone)] // The parser should handle multiple derive macros.
// #[derive(Configurable)] // This is what a real use-case would look like.
pub struct ServerConfig {
    // Field-level attributes are a key challenge for parsers.
    // #[config(default = "127.0.0.1")]
    pub host: String,

    // #[config(validate_with = "validate_port")]
    pub port: u16,
    
    /// Private configuration field
    #[allow(dead_code)]
    internal_key: Option<String>,
}

// A dummy implementation to make the code runnable.
// A proc-macro would generate this automatically.
impl Configurable for ServerConfig {
    fn from_source<S: DataSource>(source: &S) -> Self {
        // In a real macro, this would read from the source and parse.
        let _reader = source.get_reader();
        ServerConfig {
            host: "localhost".to_string(),
            port: 8080,
            internal_key: None,
        }
    }
}

/// A function with a Higher-Rank Trait Bound (HRTB).
/// The `F` closure must work for *any* lifetime `'a`.
pub fn process_all_sources<F>(sources: Vec<&dyn DataSource>, mut processor: F)
where
    F: for<'a> FnMut(Box<dyn std::io::Read + 'a>),
{
    for source in sources {
        let reader = source.get_reader();
        // The type of `reader` is tied to the lifetime of `source` in this loop iteration.
        // The closure `processor` must be able to handle this.
        processor(Box::new(reader));
    }
}

/// Advanced trait with const generics and GATs
pub trait AdvancedContainer<const N: usize> {
    type Item<'a>: Clone where Self: 'a;
    type Iterator<'a>: Iterator<Item = Self::Item<'a>> where Self: 'a;
    
    fn get_items<'a>(&'a self) -> Self::Iterator<'a>;
    fn process_batch<'a, F>(&'a self, f: F) -> [Option<Self::Item<'a>>; N]
    where
        F: Fn(usize) -> Option<Self::Item<'a>>;
}

/// Implementation with const generics
pub struct FixedArray<T: Clone, const N: usize> {
    data: [Option<T>; N],
}

impl<T: Clone, const N: usize> FixedArray<T, N> {
    /// Create new fixed array
    pub const fn new() -> Self {
        Self {
            data: [None; N],
        }
    }
    
    /// Private validation method
    fn is_valid_index(&self, index: usize) -> bool {
        index < N
    }
    
    /// Internal method for unsafe operations
    pub(crate) unsafe fn get_unchecked(&self, index: usize) -> Option<&T> {
        self.data.get_unchecked(index).as_ref()
    }
}

impl<T: Clone, const N: usize> AdvancedContainer<N> for FixedArray<T, N> {
    type Item<'a> = &'a T where T: 'a;
    type Iterator<'a> = std::iter::FilterMap<
        std::slice::Iter<'a, Option<T>>, 
        fn(&'a Option<T>) -> Option<&'a T>
    > where T: 'a;
    
    fn get_items<'a>(&'a self) -> Self::Iterator<'a> {
        self.data.iter().filter_map(|x| x.as_ref())
    }
    
    fn process_batch<'a, F>(&'a self, f: F) -> [Option<Self::Item<'a>>; N]
    where
        F: Fn(usize) -> Option<Self::Item<'a>>
    {
        let mut result: [Option<Self::Item<'a>>; N] = [None; N];
        for i in 0..N {
            result[i] = f(i);
        }
        result
    }
}

/// Advanced async trait with GATs
pub trait AsyncDataProcessor {
    type Output<'a>: Send where Self: 'a;
    type Error: std::error::Error + Send + Sync;
    
    async fn process_async<'a>(&'a self, data: &'a [u8]) -> Result<Self::Output<'a>, Self::Error>;
    
    /// Default implementation with complex bounds
    async fn batch_process<'a, I>(&'a self, inputs: I) -> Vec<Result<Self::Output<'a>, Self::Error>>
    where
        I: IntoIterator<Item = &'a [u8]> + Send,
        I::IntoIter: Send,
    {
        let mut results = Vec::new();
        for input in inputs {
            results.push(self.process_async(input).await);
        }
        results
    }
}

fn main() {
    struct FileSource { path: String }
    impl DataSource for FileSource {
        type Reader<'a> = std::io::Cursor<&'a [u8]>;
        type Config<T: Clone> = T;
        
        fn get_reader<'a>(&'a self) -> Self::Reader<'a> {
            // Dummy implementation
            std::io::Cursor::new(self.path.as_bytes())
        }
        
        fn get_config<T: Clone>(&self) -> Self::Config<T> {
            T::default()
        }
        
        fn process_with_bounds<'a, T, U>(&'a self, input: T) -> U 
        where
            T: AsRef<str> + 'a,
            U: From<&'a str> + Default,
            Self::Reader<'a>: std::io::BufRead,
        {
            U::from(input.as_ref())
        }
    }

    let file_source = FileSource { path: "config.toml".to_string() };
    
    println!("Processing sources with a complex generic function.");
    process_all_sources(vec![&file_source], |mut reader| {
        let mut content = String::new();
        use std::io::Read;
        let _ = reader.read_to_string(&mut content);
        println!("Read from source: {}", content);
    });
    
    // Test const generics
    let array: FixedArray<String, 5> = FixedArray::new();
    let _ = array.get_items();
}