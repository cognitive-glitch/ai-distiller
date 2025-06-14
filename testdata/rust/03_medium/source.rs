// 03_medium.rs
// A test for generics, lifetimes, advanced traits, and error handling.
use std::fmt::{Debug, Display};

/// A custom error type for our parsing operations.
// The parser should handle enum definitions and derive attributes.
#[derive(Debug)]
pub enum AnalysisError {
    IoError(std::io::Error),
    EmptyContent,
    InvalidFormat(String),
}

/// A trait for a data source that can be analyzed.
/// This uses an associated type, a more advanced trait feature.
pub trait DataSource {
    type Content: AsRef<[u8]>;
    fn get_content(&self) -> Result<Self::Content, AnalysisError>;
    
    /// Default method with lifetime parameters
    fn content_slice<'a>(&'a self) -> Option<&'a [u8]> where Self::Content: 'a {
        None // Default implementation
    }
}

/// A generic container for an analysis result.
/// It's generic over the type `T` which must implement `Display`.
pub struct AnalysisResult<T: Display> {
    source_id: String,
    result: T,
    metadata: Option<String>,
}

impl<T: Display> AnalysisResult<T> {
    /// Create new analysis result
    pub fn new(source_id: String, result: T) -> Self {
        Self {
            source_id,
            result,
            metadata: None,
        }
    }
    
    /// Private validation method
    fn is_valid(&self) -> bool {
        !self.source_id.is_empty()
    }
    
    /// Internal metadata setter
    pub(crate) fn set_metadata(&mut self, metadata: String) {
        self.metadata = Some(metadata);
    }
}

/// A generic function with a lifetime `'a` and trait bounds.
/// It analyzes a data source and returns a result.
// The parser must correctly handle lifetimes and `where` clauses.
pub fn analyze<'a, S>(source: &'a S) -> Result<AnalysisResult<String>, AnalysisError>
where
    S: DataSource + ?Sized, // `?Sized` is an interesting bound to parse.
{
    let content = source.get_content()?.as_ref().to_vec();

    if content.is_empty() {
        return Err(AnalysisError::EmptyContent);
    }

    // A mock analysis.
    let analysis_summary = format!("Analyzed {} bytes", content.len());

    Ok(AnalysisResult::new("mock_id".to_string(), analysis_summary))
}

/// Advanced generic function with multiple lifetime parameters
pub fn compare_sources<'a, 'b, S1, S2>(
    source1: &'a S1, 
    source2: &'b S2
) -> Result<bool, AnalysisError>
where
    S1: DataSource + Debug,
    S2: DataSource + Debug,
{
    let content1 = source1.get_content()?;
    let content2 = source2.get_content()?;
    
    Ok(content1.as_ref() == content2.as_ref())
}

// An example implementation of our DataSource.
struct InMemorySource {
    data: Vec<u8>,
}

impl DataSource for InMemorySource {
    type Content = Vec<u8>; // The associated type is specified here.
    fn get_content(&self) -> Result<Self::Content, AnalysisError> {
        Ok(self.data.clone())
    }
}

impl Debug for InMemorySource {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "InMemorySource {{ {} bytes }}", self.data.len())
    }
}

/// Generic trait with lifetime bounds
pub trait Processor<'a, T> 
where 
    T: Clone + 'a 
{
    type Output: 'a;
    
    fn process(&self, input: &'a T) -> Self::Output;
    
    /// Private helper method
    fn validate_input(&self, _input: &T) -> bool {
        true
    }
}

/// Implementation for string processing
pub struct StringProcessor;

impl<'a> Processor<'a, String> for StringProcessor {
    type Output = &'a str;
    
    fn process(&self, input: &'a String) -> Self::Output {
        if self.validate_input(input) {
            input.as_str()
        } else {
            ""
        }
    }
}

fn main() {
    let source = InMemorySource { data: vec![1, 2, 3] };
    match analyze(&source) {
        Ok(res) => println!("Success from {}: {}", res.source_id, res.result),
        Err(e) => println!("Error: {:?}", e),
    }
}