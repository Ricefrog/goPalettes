import React, { useState } from 'react';

const PaletteExtractForm = () => {
	const [numCols, setNumCols] = useState(3)
	const [selectedFile, setSelectedFile] = useState(null)

	const options = []
	for (let i = 3; i < 11; i++) {
		options.push(<option value={i} key={i}>{i}</option>)
	}

	return (
		<form onSubmit={() => console.log("SUBMIT")}>
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
				value={selectedFile}
				onChange={(e) => setSelectedFile(e.target.files[0])}
			/>
			<button>
				Find palette!
			</button>
		</form>
	)
}

export default PaletteExtractForm;
