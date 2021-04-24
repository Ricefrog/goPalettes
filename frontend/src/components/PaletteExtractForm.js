import React, { useState } from 'react';

const API_URL = 'http://localhost:8080/api';

const PaletteExtractForm = () => {
	const [numCols, setNumCols] = useState(3)
	const [selectedFile, setSelectedFile] = useState(undefined)

	const options = []
	for (let i = 3; i < 11; i++) {
		options.push(<option value={i} key={i}>{i}</option>)
	}

	const submitForm = () => {
		console.log('Uploading extract form.');
		const formData = new FormData();
		formData.append('image', selectedFile)
		formData.append('numOfColors', numCols)

		const options = {
			method: 'POST',
			headers: {
				'Content-Type': 'multipart/form-data',
			},
			body: formData,
		}
		delete options.headers['Content-Type'];

		fetch(API_URL+'/extract', options)
		.then(response => response.json())
		.then(result => {
			console.log('Success:', result);
		})
		.catch(error => {
			console.error('Error:', error);
		});
	}

	return (
		<form onSubmit={submitForm}>
			<div>Extract a palette from a file.</div>
			<label htmlFor="extractNum">Colors to extract:</label>
			<select 
				id="extractNum"
				value={numCols}
				onChange={(e) => setNumCols(e.target.value)}
			>
				{options}
			</select>
			<input 
				type="file" 
				name="image"
				onChange={(e) => setSelectedFile(e.target.files[0])}
			/>
			<button>
				Find palette!
			</button>
		</form>
	)
}

export default PaletteExtractForm;
